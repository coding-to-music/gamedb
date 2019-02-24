package queue

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/cenkalti/backoff"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/websockets"
	"github.com/gocolly/colly"
	influx "github.com/influxdata/influxdb1-client"
	"github.com/mitchellh/mapstructure"
	"github.com/streadway/amqp"
)

type bundleMessage struct {
	ID    int `json:"id"`
	AppID int `json:"app_id"`
}

type bundleQueue struct {
	baseQueue
}

func (q bundleQueue) processMessages(msgs []amqp.Delivery) {

	msg := msgs[0]

	var err error
	var payload = baseMessage{
		Message: bundleMessage{},
	}

	err = helpers.Unmarshal(msg.Body, &payload)
	if err != nil {
		logError(err)
		payload.ack(msg)
		return
	}

	var message bundleMessage
	err = mapstructure.Decode(payload.Message, &message)
	if err != nil {
		logError(err)
		payload.ack(msg)
		return
	}

	if payload.Attempt > 1 {
		logInfo("Consuming bundle " + strconv.Itoa(message.ID) + ", attempt " + strconv.Itoa(payload.Attempt))
	}

	// Load current bundle
	gorm, err := db.GetMySQLClient()
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	bundle := db.Bundle{}
	gorm = gorm.FirstOrInit(&bundle, db.Bundle{ID: message.ID})
	if gorm.Error != nil {
		logError(gorm.Error, message.ID)
		payload.ackRetry(msg)
		return
	}

	appIDs, err := bundle.GetAppIDs()
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	if message.AppID > 0 && helpers.SliceHasInt(appIDs, message.AppID) {
		logInfo("Skipping, bundle already has app")
		payload.ack(msg)
		return
	}

	err = updateBundle(&bundle)
	if err != nil && err != steam.ErrAppNotFound {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	// Save new data
	gorm = gorm.Save(&bundle)
	if gorm.Error != nil {
		logError(gorm.Error, message.ID)
		payload.ackRetry(msg)
		return
	}

	// Save to InfluxDB
	err = saveBundleToInflux(bundle)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	// Send websocket
	page, err := websockets.GetPage(websockets.PageBundle)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	} else if page.HasConnections() {
		page.Send(message.ID)
	}

	page, err = websockets.GetPage(websockets.PageBundles)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	if page.HasConnections() {
		page.Send(message.ID)
	}

	payload.ack(msg)
}

func updateBundle(bundle *db.Bundle) (err error) {

	var apps []string
	var packages []string

	c := colly.NewCollector(
		colly.AllowedDomains("store.steampowered.com"),
		colly.AllowURLRevisit(), // This is for retrys
	)

	// Set cookies
	cookieURL, _ := url.Parse("https://store.steampowered.com")

	cookieJar, err := cookiejar.New(nil)
	cookieJar.SetCookies(cookieURL, []*http.Cookie{
		{Name: "birthtime", Value: "536457601", Path: "/", Domain: "store.steampowered.com"},
		{Name: "lastagecheckage", Value: "1-January-1987", Path: "/", Domain: "store.steampowered.com"},
	})

	c.SetCookieJar(cookieJar)

	// Title
	c.OnHTML("h2.pageheader", func(e *colly.HTMLElement) {
		bundle.Name = e.Text
	})

	// Image
	c.OnHTML("img.package_header", func(e *colly.HTMLElement) {
		bundle.Image = e.Attr("src")
	})

	// Discount
	c.OnHTML(".game_purchase_discount .bundle_base_discount", func(e *colly.HTMLElement) {
		bundle.Discount, err = strconv.Atoi(strings.Replace(e.Text, "%", "", 1))
	})

	// Bigger discount
	c.OnHTML(".game_purchase_discount .discount_pct", func(e *colly.HTMLElement) {
		bundle.Discount, err = strconv.Atoi(strings.Replace(e.Text, "%", "", 1))
	})

	// Apps
	c.OnHTML("[data-ds-appid]", func(e *colly.HTMLElement) {
		apps = append(apps, strings.Split(e.Attr("data-ds-appid"), ",")...)
	})

	// Packages
	c.OnHTML("[data-ds-packageid]", func(e *colly.HTMLElement) {
		packages = append(packages, strings.Split(e.Attr("data-ds-packageid"), ",")...)
	})

	// Retry call
	operation := func() (err error) {
		return c.Visit("https://store.steampowered.com/bundle/" + strconv.Itoa(bundle.ID))
	}

	policy := backoff.NewExponentialBackOff()

	err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { logInfo(err) })
	if err != nil {
		return err
	}

	//
	if len(apps) == 0 && len(packages) == 0 {
		return nil
	}

	// Apps
	b, err := json.Marshal(helpers.StringSliceToIntSlice(apps))
	if err != nil {
		return err
	}

	bundle.AppIDs = string(b)

	// Packages
	b, err = json.Marshal(helpers.StringSliceToIntSlice(packages))
	if err != nil {
		return err
	}

	bundle.PackageIDs = string(b)

	return nil
}

func saveBundleToInflux(bundle db.Bundle) (err error) {

	_, err = db.InfluxWrite(db.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(db.InfluxMeasurementApps),
		Tags: map[string]string{
			"bundle_id": strconv.Itoa(bundle.ID),
		},
		Fields: map[string]interface{}{
			"discount": bundle.Discount,
		},
		Time:      time.Now(),
		Precision: "m",
	})

	return err
}
