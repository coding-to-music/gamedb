package queue

import (
	"encoding/json"
	"math"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/Jleagle/unmarshal-go/ctypes"
	"github.com/cenkalti/backoff/v4"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/steam"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/gocolly/colly/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type BundleMessage struct {
	ID int `json:"id"`
}

func bundleHandler(message *rabbit.Message) {

	payload := BundleMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	// Load current bundle
	var newBundle bool

	bundle, err := mongo.GetBundle(payload.ID)
	if err == mongo.ErrNoDocuments {

		bundle = mongo.Bundle{}
		bundle.ID = payload.ID
		bundle.CreatedAt = time.Now()

		newBundle = true

	} else if err != nil {

		log.Err("GetBundle", zap.Error(err), zap.Int("bundle", payload.ID))
		sendToRetryQueue(message)
		return
	}

	oldBundle := bundle

	err = updateBundle(&bundle)
	if err != nil && err != steamapi.ErrAppNotFound {
		steam.LogSteamError(err, zap.Int("bundle id", payload.ID))
		sendToRetryQueue(message)
		return
	}

	var wg sync.WaitGroup

	// Save new data
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		_, err = mongo.ReplaceOne(mongo.CollectionBundles, bson.D{{"_id", bundle.ID}}, bundle)
		if err != nil {
			log.ErrS(err, payload.ID)
			sendToRetryQueue(message)
			return
		}
	}()

	// Save price change to Mongo
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err = saveBundlePriceToMongo(bundle, oldBundle)
		if err != nil {
			log.ErrS(err, payload.ID)
			sendToRetryQueue(message)
			return
		}
	}()

	wg.Wait()

	if message.ActionTaken {
		return
	}

	// Clear caches
	items := []string{
		memcache.ItemBundleInQueue(bundle.ID).Key,
		memcache.ItemBundle(bundle.ID).Key,
	}

	err = memcache.Delete(items...)
	if err != nil {
		log.ErrS(err, payload.ID)
		sendToRetryQueue(message)
		return
	}

	// Send websocket
	if newBundle {

		wsPayload := IntPayload{ID: payload.ID}
		err = ProduceWebsocket(wsPayload, websockets.PageBundle, websockets.PageBundles)
		if err != nil {
			log.ErrS(err, payload.ID)
			sendToRetryQueue(message)
			return
		}
	}

	// Elastic
	err = ProduceBundleSearch(bundle)
	if err != nil {
		log.Err("Producing bundle search", zap.Error(err), zap.Int("bundle", payload.ID))
		sendToRetryQueue(message)
		return
	}

	message.Ack()
}

func updateBundle(bundle *mongo.Bundle) (err error) {

	var prices = map[steamapi.ProductCC]int{}
	var pricesSale = map[steamapi.ProductCC]int{}

	for _, prodCC := range i18n.GetProdCCs(true) {

		c := colly.NewCollector(
			colly.AllowedDomains("store.steampowered.com"),
			colly.AllowURLRevisit(),
			steam.WithAgeCheckCookie,
			steam.WithTimeout(0),
		)

		if prodCC.ProductCode == steamapi.ProductCCUS {

			var apps []int
			var packages []int

			// Title
			c.OnHTML("h2.pageheader", func(e *colly.HTMLElement) {
				bundle.Name = strings.TrimSpace(e.Text)
			})

			// Image
			c.OnHTML("img.package_header", func(e *colly.HTMLElement) {
				bundle.Image = e.Attr("src")
			})

			// Sale discount
			c.OnHTML(".game_purchase_discount .discount_pct", func(e *colly.HTMLElement) {

				bundle.DiscountSale, err = strconv.Atoi(strings.Replace(e.Text, "%", "", 1))
				if err != nil {
					log.ErrS(err)
				}
			})

			// Bundle data
			c.OnHTML("[data-ds-bundle-data]", func(e *colly.HTMLElement) {

				var data bundleData
				err = json.Unmarshal([]byte(e.Attr("data-ds-bundle-data")), &data)
				if err != nil {
					log.Err("Reading bundle data", zap.Error(err), zap.Int("bundle", bundle.ID))
				}

				bundle.Giftable = !data.RestrictGifting
				bundle.Discount = int(data.DiscountPct)

				if data.MustPurchaseAsSet {
					bundle.Type = mongo.BundleTypePurchaseTogether
				} else {
					bundle.Type = mongo.BundleTypeCompleteTheSet
				}

				for _, v := range data.Items {
					apps = append(apps, v.IncludedAppIDs...)
					packages = append(packages, v.PackageID)
				}

				bundle.Apps = apps
				bundle.Packages = packages
			})
		}

		// Price
		c.OnHTML(".game_purchase_discount .discount_original_price", func(e *colly.HTMLElement) {

			val := helpers.RegexNonInts.ReplaceAllString(e.Text, "")
			if len(val) > 0 {
				i, _ := strconv.Atoi(val)
				prices[prodCC.ProductCode] = i
			}
		})

		c.OnHTML(".game_purchase_discount .discount_final_price", func(e *colly.HTMLElement) {

			val := helpers.RegexNonInts.ReplaceAllString(e.Text, "")
			if len(val) > 0 {
				i, _ := strconv.Atoi(val)
				pricesSale[prodCC.ProductCode] = i
			}
		})

		// Retry call
		operation := func() (err error) {

			q := url.Values{}
			q.Set("cc", string(prodCC.ProductCode))

			return c.Visit("https://store.steampowered.com/bundle/" + strconv.Itoa(bundle.ID) + "?" + q.Encode())
		}

		policy := backoff.NewExponentialBackOff()

		err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Info("Scraping bundle", zap.Error(err)) })
		if err != nil {
			return err
		}

		// Prices
		bundle.Prices = prices
		bundle.PricesSale = pricesSale
	}

	return nil
}

type bundleData struct {
	DiscountPct       ctypes.Int  `json:"m_nDiscountPct"`
	MustPurchaseAsSet ctypes.Bool `json:"m_bMustPurchaseAsSet"`
	Items             []struct {
		PackageID                    int   `json:"m_nPackageID"`
		IncludedAppIDs               []int `json:"m_rgIncludedAppIDs"`
		PackageDiscounted            bool  `json:"m_bPackageDiscounted"`
		BasePriceInCents             int   `json:"m_nBasePriceInCents"`
		FinalPriceInCents            int   `json:"m_nFinalPriceInCents"`
		FinalPriceWithBundleDiscount int   `json:"m_nFinalPriceWithBundleDiscount"`
	} `json:"m_rgItems"`
	IsCommercial    bool `json:"m_bIsCommercial"`
	RestrictGifting bool `json:"m_bRestrictGifting"`
}

var bundlePriceLock sync.Mutex

func saveBundlePriceToMongo(bundle mongo.Bundle, oldBundle mongo.Bundle) (err error) {

	bundlePriceLock.Lock()
	defer bundlePriceLock.Unlock()

	if math.Abs(float64(bundle.DiscountSale)) != math.Abs(float64(oldBundle.DiscountSale)) {

		doc := mongo.BundlePrice{
			CreatedAt: time.Now(),
			BundleID:  bundle.ID,
			Discount:  int(math.Abs(float64(bundle.DiscountSale))),
		}

		// Does a replace, as sometimes doing a InsertOne would error on key already existing
		_, err = mongo.ReplaceOne(mongo.CollectionBundlePrices, bson.D{{"_id", doc.GetKey()}}, doc)
	}

	return err
}
