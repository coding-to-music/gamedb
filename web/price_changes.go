package web

import (
	"net/http"
	"strconv"
	"sync"

	"cloud.google.com/go/datastore"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/logging"
	"github.com/gamedb/website/session"
)

func PriceChangesHandler(w http.ResponseWriter, r *http.Request) {

	t := priceChangesTemplate{}
	t.Fill(w, r, "Price Changes")

	returnTemplate(w, r, "price_changes", t)
}

type priceChangesTemplate struct {
	GlobalTemplate
}

func PriceChangesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	query := DataTablesQuery{}
	query.FillFromURL(r.URL.Query())

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

		logging.Error(err)

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
