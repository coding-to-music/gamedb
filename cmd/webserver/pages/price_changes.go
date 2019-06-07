package pages

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

func PriceChangeRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", priceChangesHandler)
	r.Get("/price-changes.json", priceChangesAjaxHandler)
	return r
}

func priceChangesHandler(w http.ResponseWriter, r *http.Request) {

	t := priceChangesTemplate{}
	t.fill(w, r, "Price Changes", "Pick up a bargain.")
	t.addAssetChosen()
	t.addAssetSlider()

	err := returnTemplate(w, r, "price_changes", t)
	log.Err(err, r)
}

type priceChangesTemplate struct {
	GlobalTemplate
}

func priceChangesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	//
	var wg sync.WaitGroup

	// Get ranks
	var priceChanges []mongo.ProductPrice

	var code = getCountryCode(r)

	var dateLimit = time.Now().AddDate(0, 0, -30)

	var filter = mongo.D{
		{"currency", string(code)},
		{"created_at", mongo.M{"$gt": dateLimit}},
	}

	typex := query.getSearchString("type")
	if typex == "apps" {
		filter = append(filter, bson.E{Key: "app_id", Value: bson.M{"$gt": 0}})
	} else if typex == "packages" {
		filter = append(filter, bson.E{Key: "package_id", Value: bson.M{"$gt": 0}})
	}

	percents := query.getSearchSlice("percents")
	if len(percents) == 2 {
		if percents[0] != "-100.00" {
			min, err := strconv.ParseFloat(percents[0], 64)
			log.Err(err)
			if err == nil {
				filter = append(filter, bson.E{Key: "difference_percent", Value: mongo.M{"$gte": min}})
			}
		}
		if percents[1] != "100.00" {
			max, err := strconv.ParseFloat(percents[1], 64)
			log.Err(err)
			if err == nil {
				filter = append(filter, bson.E{Key: "difference_percent", Value: mongo.M{"$lte": max}})
			}
		}
	}

	prices := query.getSearchSlice("prices")
	if len(prices) == 2 {
		if prices[0] != "0.00" {
			min, err := strconv.Atoi(strings.Replace(prices[0], ".", "", 1))
			log.Err(err)
			if err == nil {
				filter = append(filter, bson.E{Key: "price_after", Value: mongo.M{"$gte": min}})
			}
		}
		if prices[1] != "100.00" {
			max, err := strconv.Atoi(strings.Replace(prices[1], ".", "", 1))
			log.Err(err)
			if err == nil {
				filter = append(filter, bson.E{Key: "price_after", Value: mongo.M{"$lte": max}})
			}
		}
	}

	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		priceChanges, err = mongo.GetPrices(query.getOffset64(), 100, filter)
		log.Err(err, r)
	}(r)

	// Get filtered
	var filtered int64
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		filtered, err = mongo.CountDocuments(mongo.CollectionProductPrices, filter, 0)
		log.Err(err, r)
	}(r)

	// Get total
	var total int64
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionProductPrices, mongo.M{
			"currency":   string(code),
			"created_at": mongo.M{"$gt": dateLimit},
		}, 0)
		log.Err(err, r)
	}(r)

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = total
	response.RecordsFiltered = filtered
	response.Draw = query.Draw

	for _, v := range priceChanges {

		response.AddRow(v.OutputForJSON())
	}

	response.output(w, r)
}
