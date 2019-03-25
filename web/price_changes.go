package web

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gamedb/website/log"
	"github.com/gamedb/website/mongo"
	"github.com/gamedb/website/session"
	"github.com/go-chi/chi"
)

func priceChangeRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", priceChangesHandler)
	r.Get("/price-changes.json", priceChangesAjaxHandler)
	return r
}

func priceChangesHandler(w http.ResponseWriter, r *http.Request) {

	t := priceChangesTemplate{}
	t.fill(w, r, "Price Changes", "Pick up a bargain.")

	err := returnTemplate(w, r, "price_changes", t)
	log.Err(err, r)
}

type priceChangesTemplate struct {
	GlobalTemplate
}

func priceChangesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	setCacheHeaders(w, 0)

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	//
	var wg sync.WaitGroup

	// Get ranks
	var priceChanges []mongo.ProductPrice

	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		priceChanges, err = mongo.GetPrices(query.getOffset64(), session.GetCountryCode(r))
		log.Err(err, r)
	}(r)

	// Get total
	var total int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		total, err = mongo.CountPrices()
		log.Err(err, r)
	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.FormatInt(total, 10)
	response.RecordsFiltered = strconv.FormatInt(total, 10)
	response.Draw = query.Draw

	for _, v := range priceChanges {

		response.AddRow(v.OutputForJSON())
	}

	response.output(w, r)
}
