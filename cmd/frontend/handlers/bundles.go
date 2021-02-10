package handlers

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi"
	"github.com/olivere/elastic/v7"
	"go.uber.org/zap"
)

func BundlesRouter() http.Handler {

	r := chi.NewRouter()
	r.Mount("/{id}", BundleRouter())

	r.Get("/", bundlesHandler)
	r.Get("/bundles.json", bundlesAjaxHandler)

	return r
}

func bundlesHandler(w http.ResponseWriter, r *http.Request) {

	t := bundlesTemplate{}
	t.fill(w, r, "bundles", "Bundles", "All the bundles on Steam")
	t.addAssetChosen()
	t.addAssetSlider()

	returnTemplate(w, r, t)
}

type bundlesTemplate struct {
	globalTemplate
}

func bundlesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, false)

	//
	var wg sync.WaitGroup

	// Get apps
	var bundles []elasticsearch.Bundle
	var countFiltered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var code = session.GetProductCC(r)
		var err error
		var sortCols = map[string]string{
			"1": "sale_discount",
			"2": "sale_prices." + string(code),
			"3": "apps",
			"4": "packages",
			"5": "updated_at",
		}

		var filters []elastic.Query

		//
		typex := query.GetSearchString("type")
		switch typex {
		case "cts", "pt":
			filters = append(filters, elastic.NewTermQuery("type", typex))
		}

		//
		giftable := query.GetSearchString("giftable")
		switch giftable {
		case "1":
			filters = append(filters, elastic.NewTermQuery("giftable", true))
		}

		//
		onsale := query.GetSearchString("onsale")
		switch onsale {
		case "1":
			filters = append(filters, elastic.NewTermQuery("on_sale", true))
		}

		//
		discount := query.GetSearchSlice("discount")
		if len(discount) == 2 {
			if discount[0] != "0" {
				min, err := strconv.Atoi(discount[0])
				if err == nil {
					filters = append(filters, elastic.NewRangeQuery("discount").Gte(min))
				}
			}
			if discount[1] != "100" {
				max, err := strconv.Atoi(discount[1])
				if err == nil {
					filters = append(filters, elastic.NewRangeQuery("discount").Lte(max))
				}
			}
		}

		//
		apps := query.GetSearchSlice("apps")
		if len(apps) == 2 {

			if apps[0] != "0" {
				min, err := strconv.Atoi(apps[0])
				if err == nil {
					filters = append(filters, elastic.NewRangeQuery("apps").Gte(min))
				}
			}

			if apps[1] != "100" {
				max, err := strconv.Atoi(apps[1])
				if err == nil {
					filters = append(filters, elastic.NewRangeQuery("apps").Lte(max))
				}
			}
		}

		//
		packages := query.GetSearchSlice("packages")
		if len(packages) == 2 {
			if packages[0] != "0" {
				min, err := strconv.Atoi(packages[0])
				if err == nil {
					filters = append(filters, elastic.NewRangeQuery("packages").Gte(min))
				}
			}
			if packages[1] != "100" {
				max, err := strconv.Atoi(packages[1])
				if err == nil {
					filters = append(filters, elastic.NewRangeQuery("packages").Lte(max))
				}
			}
		}

		filter := elastic.NewBoolQuery().Filter(filters...)

		bundles, countFiltered, err = elasticsearch.SearchBundles(query.GetOffset(), 100, "", query.GetOrderElastic(sortCols), filter)
		if err != nil {
			log.Err("Searching bundles", zap.Error(err))
		}
	}()

	// Get total
	var count int
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mysql.CountBundles()
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, int64(count), countFiltered, nil)
	for _, v := range bundles {
		response.AddRow(v.OutputForJSON())
	}

	returnJSON(w, r, response)
}
