package pages

import (
	"net/http"
	"strconv"
	"strings"
	"sync"

	webserverHelpers "github.com/gamedb/gamedb/cmd/webserver/helpers"
	"github.com/gamedb/gamedb/cmd/webserver/helpers/datatable"
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
	t.fill(w, r, "Price Changes", "All game price changes")
	t.addAssetChosen()
	t.addAssetSlider()

	returnTemplate(w, r, "price_changes", t)
}

type priceChangesTemplate struct {
	GlobalTemplate
}

func priceChangesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, true)

	//
	var wg sync.WaitGroup

	// Get ranks
	var code = webserverHelpers.GetProductCC(r)

	var filter = bson.D{
		{Key: "prod_cc", Value: string(code)},
	}

	typex := query.GetSearchString("type")
	if typex == "apps" {
		filter = append(filter, bson.E{Key: "app_id", Value: bson.M{"$gt": 0}})
	} else if typex == "packages" {
		filter = append(filter, bson.E{Key: "package_id", Value: bson.M{"$gt": 0}})
	}

	percents := query.GetSearchSlice("change")
	if len(percents) == 2 {
		if percents[0] != "-100.00" {
			min, err := strconv.ParseFloat(percents[0], 64)
			log.Err(err)
			if err == nil {
				filter = append(filter, bson.E{Key: "difference_percent", Value: bson.M{"$gte": min}})
			}

			// Dont show infinite difference_percent
			if min > -100 {
				filter = append(filter, bson.E{Key: "$or", Value: bson.A{
					bson.M{"difference_percent": bson.M{"$gt": 0}},
					bson.M{"difference_percent": bson.M{"$lt": 0}},
					bson.M{"difference": bson.M{"$gte": 0}},
				}})
			}
		}
		if percents[1] != "100.00" {
			max, err := strconv.ParseFloat(percents[1], 64)
			log.Err(err)
			if err == nil {
				filter = append(filter, bson.E{Key: "difference_percent", Value: bson.M{"$lte": max}})
			}

			// Dont show infinite difference_percent
			if max < 100 {
				filter = append(filter, bson.E{Key: "$or", Value: bson.A{
					bson.M{"difference_percent": bson.M{"$gt": 0}},
					bson.M{"difference_percent": bson.M{"$lt": 0}},
					bson.M{"difference": bson.M{"$lte": 0}},
				}})
			}
		}
	}

	prices := query.GetSearchSlice("price")
	if len(prices) == 2 {
		if prices[0] != "0.00" {
			min, err := strconv.Atoi(strings.Replace(prices[0], ".", "", 1))
			log.Err(err)
			if err == nil {
				filter = append(filter, bson.E{Key: "price_after", Value: bson.M{"$gte": min}})
			}
		}
		if prices[1] != "100.00" {
			max, err := strconv.Atoi(strings.Replace(prices[1], ".", "", 1))
			log.Err(err)
			if err == nil {
				filter = append(filter, bson.E{Key: "price_after", Value: bson.M{"$lte": max}})
			}
		}
	}

	// Get rows
	var priceChanges []mongo.ProductPrice
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		priceChanges, err = mongo.GetPrices(query.GetOffset64(), 100, filter)
		if err != nil {
			log.Err(err, r)
			return
		}
	}(r)

	// Get filtered count
	var filtered int64
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		filtered, err = mongo.CountDocuments(mongo.CollectionProductPrices, filter, 0)
		log.Err(err, r)
	}(r)

	// Get total count
	var total int64
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionProductPrices, bson.D{{Key: "prod_cc", Value: string(code)}}, 0)
		log.Err(err, r)
	}(r)

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, total, filtered)
	for _, price := range priceChanges {

		response.AddRow(price.OutputForJSON())
	}

	returnJSON(w, r, response)
}
