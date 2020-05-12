package queue

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/cenkalti/backoff/v4"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	steamHelper "github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/gocolly/colly"
	"go.mongodb.org/mongo-driver/bson"
)

type BundleMessage struct {
	ID int `json:"id"`
}

func bundleHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := BundleMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		// Load current bundle
		gorm, err := sql.GetMySQLClient()
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			continue
		}

		bundle := sql.Bundle{}
		gorm = gorm.FirstOrInit(&bundle, sql.Bundle{ID: payload.ID})
		if gorm.Error != nil {
			log.Err(gorm.Error, payload.ID)
			sendToRetryQueue(message)
			continue
		}

		oldBundle := bundle

		err = updateBundle(&bundle)
		if err != nil && err != steamapi.ErrAppNotFound {
			steamHelper.LogSteamError(err, payload.ID)
			sendToRetryQueue(message)
			continue
		}

		var wg sync.WaitGroup

		// Save new data
		wg.Add(1)
		go func() {

			defer wg.Done()

			gorm = gorm.Save(&bundle)
			if gorm.Error != nil {
				log.Err(gorm.Error, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		// Save to InfluxDB
		wg.Add(1)
		go func() {

			defer wg.Done()

			var err error

			err = saveBundlePriceToMongo(bundle, oldBundle)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		wg.Wait()

		if message.ActionTaken {
			continue
		}

		// Send websocket
		wsPayload := IntPayload{ID: payload.ID}
		err = ProduceWebsocket(wsPayload, websockets.PageBundle, websockets.PageBundles)
		if err != nil {
			log.Err(err, payload.ID)
		}

		// Clear caches
		err = memcache.Delete(
			memcache.MemcacheBundleInQueue(bundle.ID).Key,
		)
		log.Err(err)

		message.Ack(false)
	}
}

func updateBundle(bundle *sql.Bundle) (err error) {

	c := colly.NewCollector(
		colly.AllowedDomains("store.steampowered.com"),
		steamHelper.WithAgeCheckCookie,
		colly.AllowURLRevisit(),
	)

	// Title
	c.OnHTML("h2.pageheader", func(e *colly.HTMLElement) {
		bundle.Name = strings.TrimSpace(e.Text)
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

	err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Info(err) })
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

var bundlePriceLock sync.Mutex

func saveBundlePriceToMongo(bundle sql.Bundle, oldBundle sql.Bundle) (err error) {

	bundlePriceLock.Lock()
	defer bundlePriceLock.Unlock()

	time.Sleep(time.Second) // prices are keyed by the second

	if bundle.Discount != oldBundle.Discount {

		doc := mongo.BundlePrice{
			CreatedAt: time.Now(),
			BundleID:  bundle.ID,
			Discount:  bundle.Discount,
		}

		// Does a replace, as sometimes doing a InsertOne would error on key already existing
		_, err = mongo.ReplaceOne(mongo.CollectionBundlePrices, bson.D{{"_id", doc.GetKey()}}, doc)
	}

	return err
}
