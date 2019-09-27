package queue

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/cenkalti/backoff"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/gocolly/colly"
	"github.com/streadway/amqp"
)

type bundleMessage struct {
	baseMessage
	Message bundleMessageInner `json:"message"`
}

type bundleMessageInner struct {
	ID    int `json:"id"`
	AppID int `json:"app_id"`
}

type bundleQueue struct {
}

func (q bundleQueue) processMessages(msgs []amqp.Delivery) {

	msg := msgs[0]

	var err error

	message := bundleMessage{}
	message.OriginalQueue = queueBundles

	err = helpers.Unmarshal(msg.Body, &message)
	if err != nil {
		logError(err, msg.Body)
		ackFail(msg, &message)
		return
	}

	if message.Attempt > 1 {
		logInfo("Consuming bundle " + strconv.Itoa(message.Message.ID) + ", attempt " + strconv.Itoa(message.Attempt))
	}

	// Load current bundle
	gorm, err := sql.GetMySQLClient()
	if err != nil {
		logError(err, message.Message.ID)
		ackRetry(msg, &message)
		return
	}

	bundle := sql.Bundle{}
	gorm = gorm.FirstOrInit(&bundle, sql.Bundle{ID: message.Message.ID})
	if gorm.Error != nil {
		logError(gorm.Error, message.Message.ID)
		ackRetry(msg, &message)
		return
	}

	oldBundle := bundle

	err = updateBundle(&bundle)
	if err != nil && err != steam.ErrAppNotFound {
		helpers.LogSteamError(err, message.Message.ID)
		ackRetry(msg, &message)
		return
	}

	var wg sync.WaitGroup

	// Save new data
	wg.Add(1)
	go func() {

		defer wg.Done()

		gorm = gorm.Save(&bundle)
		if gorm.Error != nil {
			logError(gorm.Error, message.Message.ID)
			ackRetry(msg, &message)
			return
		}
	}()

	// Save to InfluxDB
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		err = savePriceToMongo(bundle, oldBundle)
		if err != nil {
			logError(err, message.Message.ID)
			ackRetry(msg, &message)
			return
		}
	}()

	wg.Wait()

	if message.actionTaken {
		return
	}

	// Send websocket
	wsPaload := websockets.PubSubIDPayload{}
	wsPaload.ID = message.Message.ID
	wsPaload.Pages = []websockets.WebsocketPage{websockets.PageBundle, websockets.PageBundles}

	_, err = helpers.Publish(helpers.PubSubTopicWebsockets, wsPaload)
	if err != nil {
		logError(err, message.Message.ID)
	}

	message.ack(msg)
}

func updateBundle(bundle *sql.Bundle) (err error) {

	c := colly.NewCollector(
		colly.AllowedDomains("store.steampowered.com"),
		colly.AllowURLRevisit(), // This is for retrys
	)

	jar, err := helpers.GetAgeCheckCookieJar()
	if err != nil {
		return err
	}
	c.SetCookieJar(jar)

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

var priceLock sync.Mutex

func savePriceToMongo(bundle sql.Bundle, oldBundle sql.Bundle) (err error) {

	priceLock.Lock()
	defer priceLock.Unlock()

	time.Sleep(time.Second) // prices are keyed by the second

	if bundle.Discount != oldBundle.Discount {

		_, err = mongo.InsertDocument(mongo.CollectionBundlePrices, mongo.BundlePrice{
			CreatedAt: time.Now(),
			BundleID:  bundle.ID,
			Discount:  bundle.Discount,
		})
	}

	return err
}
