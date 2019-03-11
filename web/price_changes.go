package web

import (
	"net/http"
	"strconv"
	"sync"

	"cloud.google.com/go/datastore"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
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

	setNoCacheHeaders(w)

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	//
	var wg sync.WaitGroup

	// Get ranks
	var priceChanges []db.ProductPrice

	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		client, ctx, err := db.GetDSClient()
		if err != nil {

			log.Err(err, r)
			return
		}

		q := datastore.NewQuery(db.KindProductPrice).Order("-created_at").Limit(100).Offset(query.getOffset())
		q = q.Filter("currency =", string(session.GetCountryCode(r)))

		_, err = client.GetAll(ctx, q, &priceChanges)
		log.Err(err, r)

	}(r)

	// Get total
	var total int
	wg.Add(1)
	go func() {

		defer wg.Done()

		total = 10000

	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(total)
	response.RecordsFiltered = strconv.Itoa(total)
	response.Draw = query.Draw

	for _, v := range priceChanges {

		response.AddRow(v.OutputForJSON())
	}

	response.output(w, r)
}
