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
	r.Get("/ajax", priceChangesAjaxHandler)
	return r
}

func priceChangesHandler(w http.ResponseWriter, r *http.Request) {

	t := priceChangesTemplate{}
	t.Fill(w, r, "Price Changes")
	t.Description = "Pick up a bargain"

	err := returnTemplate(w, r, "price_changes", t)
	log.Log(err)
}

type priceChangesTemplate struct {
	GlobalTemplate
}

func priceChangesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	query := DataTablesQuery{}
	err := query.FillFromURL(r.URL.Query())
	log.Log(err)

	//
	var wg sync.WaitGroup

	// Get ranks
	var priceChanges []db.ProductPrice

	wg.Add(1)
	go func(r *http.Request) {

		client, ctx, err := db.GetDSClient()
		if err == nil {

			q := datastore.NewQuery(db.KindProductPrice).Limit(100).Order("-created_at")
			q = q.Filter("currency =", string(session.GetCountryCode(r)))

			q, err = query.SetOffsetDS(q)
			if err == nil {
				_, err = client.GetAll(ctx, q, &priceChanges)
			}
		}

		log.Log(err)

		wg.Done()
	}(r)

	// Get total
	var total int
	wg.Add(1)
	go func() {

		total = 10000

		wg.Done()
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

	response.output(w)
}
