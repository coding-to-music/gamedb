package web

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

const (
	PriceChangeLimit = 100
)

func PriceChangesHandler(w http.ResponseWriter, r *http.Request) {

	// Get page number
	page, err := strconv.Atoi(r.URL.Query().Get("p"))
	if err != nil {
		page = 1
	}

	var wg sync.WaitGroup

	// Get total changes
	var total int
	wg.Add(1)
	go func() {

		total, err = datastore.CountPrices()
		if err != nil {
			logger.Error(err)
		}

		wg.Done()

	}()

	// Get changes
	var changes []datastore.Price
	wg.Add(1)
	go func() {

		changes, err = datastore.GetLatestPrices(PriceChangeLimit, page)
		if err != nil {
			logger.Error(err)
			returnErrorTemplate(w, r, 500, err.Error())
			return
		}

		wg.Done()

	}()

	// Wait
	wg.Wait()

	// Template
	template := priceChangesTemplate{}
	template.Fill(r, "Price Changes")
	template.Changes = changes
	template.Pagination = Pagination{
		path:  "/price-changes?p=",
		page:  page,
		limit: PriceChangeLimit,
		total: total,
	}

	returnTemplate(w, r, "price_changes", template)
	return
}

type priceChangesTemplate struct {
	GlobalTemplate
	Apps       []mysql.App
	Changes    []datastore.Price
	Pagination Pagination
}
