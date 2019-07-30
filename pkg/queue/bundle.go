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
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/gocolly/colly"
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
		Message:       bundleMessage{},
		OriginalQueue: queueGoBundles,
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
	gorm, err := sql.GetMySQLClient()
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	bundle := sql.Bundle{}
	gorm = gorm.FirstOrInit(&bundle, sql.Bundle{ID: message.ID})
	if gorm.Error != nil {
		logError(gorm.Error, message.ID)
		payload.ackRetry(msg)
		return
	}

	oldBundle := bundle

	err = updateBundle(&bundle)
	if err != nil && err != steam.ErrAppNotFound {
		helpers.LogSteamError(err, message.ID)
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
	err = savePriceToMongo(bundle, oldBundle)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	// Send websocket
	wsPaload := websockets.PubSubIDPayload{}
	wsPaload.ID = message.ID
	wsPaload.Pages = []websockets.WebsocketPage{websockets.PageBundle, websockets.PageBundles}

	_, err = helpers.Publish(helpers.PubSubTopicWebsockets, wsPaload)
	if err != nil {
		logError(err, message.ID)
	}

	payload.ack(msg)
}

func updateBundle(bundle *sql.Bundle) (err error) {

	c := colly.NewCollector(
		colly.AllowedDomains("store.steampowered.com"),
		colly.AllowURLRevisit(), // This is for retrys
	)

	// Set cookies
	cookieURL, _ := url.Parse("https://store.steampowered.com")

	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}

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
		var discount int
		discount, err = strconv.Atoi(strings.Replace(e.Text, "%", "", 1))
		bundle.SetDiscount(discount)
	})

	// Bigger discount
	c.OnHTML(".game_purchase_discount .discount_pct", func(e *colly.HTMLElement) {
		var discount int
		discount, err = strconv.Atoi(strings.Replace(e.Text, "%", "", 1))
		bundle.SetDiscount(discount)
	})

	// Apps
	var apps []string
	c.OnHTML("[data-ds-appid]", func(e *colly.HTMLElement) {
		apps = append(apps, strings.Split(e.Attr("data-ds-appid"), ",")...)
	})

	// Packages
	var packages []string
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

func savePriceToMongo(bundle sql.Bundle, oldBundle sql.Bundle) (err error) {

	if bundle.Discount != oldBundle.Discount {

		_, err = mongo.InsertDocument(mongo.CollectionBundlePrices, mongo.BundlePrice{
			CreatedAt: time.Now(),
			BundleID:  bundle.ID,
			Discount:  bundle.Discount,
		})
	}

	return err
}
