package web

import (
	"net/http"
	"strconv"
	"sync"

	"cloud.google.com/go/datastore"
	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/logger"
)

func PriceChangesHandler(w http.ResponseWriter, r *http.Request) {

	t := priceChangesTemplate{}
	t.Fill(w, r, "Price Changes")

	returnTemplate(w, r, "price_changes", t)
	return
}

type priceChangesTemplate struct {
	GlobalTemplate
}

func PriceChangesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	query.FillFromURL(r.URL.Query())

	//
	var wg sync.WaitGroup

	// Get ranks
	var priceChanges []db.PriceChange

	wg.Add(1)
	go func() {

		client, ctx, err := db.GetDSClient()
		if err != nil {

			logger.Error(err)

		} else {

			columns := map[string]string{
				//"0": "app_name",
				"1": "release_date_unix",
				"2": "price_final",
				"3": "discount",
				"4": "change",
				"5": "created_at",
			}

			q := datastore.NewQuery(db.KindPrice).Limit(100)
			q = q.Filter("currency =", "usd")
			q = q.Filter("first =", false)

			column := query.GetOrderDS(columns, false)
			if column != "" {
				q, err = query.SetOrderOffsetDS(q, columns)
				if err != nil {

					logger.Error(err)

				} else {

					_, err := client.GetAll(ctx, q, &priceChanges)
					logger.Error(err)
				}
			}
		}

		wg.Done()
	}()

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
