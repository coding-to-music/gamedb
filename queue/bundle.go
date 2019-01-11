package queue

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/websockets"
	"github.com/gocolly/colly"
	"github.com/streadway/amqp"
)

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
	)

	// Title
	c.OnHTML("h2.pageheader", func(e *colly.HTMLElement) {
		bundle.Name = e.Text
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

	//
	err = c.Visit("https://store.steampowered.com/bundle/" + strconv.Itoa(bundle.ID))
	if err != nil {
		return err
	}

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
