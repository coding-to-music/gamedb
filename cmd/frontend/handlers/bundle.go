package handlers

import (
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
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

	// todo, validate
	// if !db.IsValidAppID(idx) {
	// 	returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid bundle ID: " + id})
	// 	return
	// }

	// Get bundle
	bundle, err := mysql.GetBundle(id, nil)
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

		appIDs, err := bundle.GetAppIDs()
		if err != nil {
			log.ErrS(err)
			return
		}

		apps, err = mongo.GetAppsByID(appIDs, nil)
		if err != nil {
			log.ErrS(err)
		}

		// Queue missing apps
		if len(appIDs) != len(apps) {
			for _, v := range appIDs {
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

		packages, err = mongo.GetPackagesByID(bundle.GetPackageIDs(), bson.M{})
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

	//
	returnTemplate(w, r, t)
}

type bundleTemplate struct {
	globalTemplate
	Bundle   mysql.Bundle
	Apps     []mongo.App
	Packages []mongo.Package
}

func bundlePricesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Err("invalid id")
		return
	}

	// Get prices
	pricesResp, err := mongo.GetBundlePrices(id)
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
	price, err := mysql.GetBundle(id, []string{"discount"})
	if err != nil {
		log.ErrS(err)
	} else {
		prices = append(prices, []int64{time.Now().Unix() * 1000, int64(price.Discount)})
	}

	// Sort prices for Highcharts
	sort.Slice(prices, func(i, j int) bool {
		return prices[i][0] < prices[j][0]
	})

	// Return
	returnJSON(w, r, prices)
}
