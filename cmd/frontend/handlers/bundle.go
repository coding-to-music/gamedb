package handlers

import (
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/go-chi/chi"
)

func BundleRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", bundleHandler)
	r.Get("/prices.json", bundlePricesAjaxHandler)
	r.Get("/{slug}", bundleHandler)
	return r
}

func bundleHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid bundle ID"})
		return
	}

	// Get bundle
	bundle, err := mongo.GetBundle(id)
	if err != nil {

		if err == mysql.ErrRecordNotFound {
			returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Sorry but we can not find this bundle."})
			return
		}

		log.ErrS(err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the bundle."})
		return
	}

	//
	var wg sync.WaitGroup

	// Get apps
	var apps []mongo.App
	wg.Add(1)
	go func() {

		defer wg.Done()

		apps, err = mongo.GetAppsByID(bundle.Apps, nil)
		if err != nil {
			log.ErrS(err)
		}

		// Queue missing apps
		if len(bundle.Apps) != len(apps) {
			for _, v := range bundle.Apps {
				var found = false
				for _, vv := range apps {
					if v == vv.ID {
						found = true
						break
					}
				}

				if !found {
					err = queue.ProduceSteam(queue.SteamMessage{AppIDs: []int{v}})
					err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
					if err != nil {
						log.ErrS(err)
					}
				}
			}
		}
	}()

	// Get packages
	var packages []mongo.Package
	wg.Add(1)
	go func() {

		defer wg.Done()

		packages, err = mongo.GetPackagesByID(bundle.Packages, nil)
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Wait
	wg.Wait()

	// Template
	t := bundleTemplate{}
	for _, v := range apps {
		if v.Background != "" {
			t.setBackground(v, true, true)
			break
		}
	}
	t.fill(w, r, "bundle", bundle.Name, "Steam bundle")
	t.addAssetHighCharts()
	t.Bundle = bundle
	t.Canonical = bundle.GetPath()
	t.Apps = apps
	t.Packages = packages

	if val, ok := bundle.GetPricesFormatted()[session.GetProductCC(r)]; ok {
		t.Price = val
	} else {
		t.Price = "-"
	}

	if val, ok := bundle.GetPricesSaleFormatted()[session.GetProductCC(r)]; ok {
		t.PriceSale = val
	} else {
		t.PriceSale = "-"
	}

	//
	returnTemplate(w, r, t)
}

type bundleTemplate struct {
	globalTemplate
	Bundle    mongo.Bundle
	Apps      []mongo.App
	Packages  []mongo.Package
	Price     string
	PriceSale string
}

func bundlePricesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Err("invalid id")
		return
	}

	// Get prices
	pricesResp, err := mongo.GetBundlePricesByID(id)
	if err != nil {
		log.ErrS(err)
		return
	}

	// Make JSON response
	var prices [][]int64

	for _, v := range pricesResp {
		prices = append(prices, []int64{v.CreatedAt.Unix() * 1000, int64(v.Discount)})
	}

	// Add current price
	price, err := mongo.GetBundle(id)
	if err != nil {
		log.ErrS(err)
	} else {
		prices = append(prices, []int64{time.Now().Unix() * 1000, int64(price.DiscountSale)})
	}

	if len(prices) == 1 {
		prices = append(prices, []int64{prices[0][0] - 1, prices[0][1]})
	}

	// Sort prices for Highcharts
	sort.Slice(prices, func(i, j int) bool {
		return prices[i][0] < prices[j][0]
	})

	// Return
	returnJSON(w, r, prices)
}
