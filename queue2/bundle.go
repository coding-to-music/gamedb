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
	"github.com/streadway/amqp"
)

func QueueBundle(bundleID int) (err error) {

	b, err := json.Marshal(RabbitMessageBundle{
		BundleID: bundleID,
	})

	return Produce(QueueBundlesData, b)
}

type RabbitMessageBundle struct {
	BundleID int
	AppID    int // The app that triggered a bundle update
}

func (d RabbitMessageBundle) getConsumeQueue() RabbitQueue {
	return QueueBundlesData
}

func (d RabbitMessageBundle) getProduceQueue() RabbitQueue {
	return QueueBundlesData
}

func (d RabbitMessageBundle) getRetryData() RabbitMessageDelay {
	return RabbitMessageDelay{}
}

func (d RabbitMessageBundle) process(msg amqp.Delivery) (requeue bool, err error) {

	// Get message payload
	message := RabbitMessageBundle{}

	err = helpers.Unmarshal(msg.Body, &message)
	if err != nil {
		return false, err
	}

	logInfo("Consuming bundle: " + strconv.Itoa(message.BundleID))

	// Load current bundle
	gorm, err := db.GetMySQLClient()
	if err != nil {
		return true, err
	}

	bundle := db.Bundle{}
	gorm = gorm.FirstOrInit(&bundle, db.Bundle{ID: message.BundleID})
	if gorm.Error != nil {
		return true, gorm.Error
	}

	appIDs, err := bundle.GetAppIDs()
	if err != nil {
		return true, err
	}

	if helpers.SliceHasInt(appIDs, message.AppID) {
		logInfo("Skipping, bundle already has app")
		return false, nil
	}

	err = updateBundle(&bundle)
	if err != nil && err != steam.ErrAppNotFound {
		return true, err
	}

	// Save new data
	gorm = gorm.Save(&bundle)
	if gorm.Error != nil {
		return true, gorm.Error
	}

	// Send websocket
	page, err := websockets.GetPage(websockets.PageBundle)
	if err != nil {
		return true, err
	} else if page.HasConnections() {
		page.Send(bundle.ID)
	}

	page, err = websockets.GetPage(websockets.PageBundles)
	if err != nil {
		return true, err
	} else if page.HasConnections() {
		page.Send(bundle.ID)
	}

	return false, nil
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
