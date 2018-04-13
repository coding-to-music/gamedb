package web

import (
	"net/http"
	"strconv"

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

	// Get total changes
	total, err := datastore.CountPrices()
	if err != nil {
		logger.Error(err)
	}

	// Get changes
	changes, err := datastore.GetAppChanges(PriceChangeLimit, page)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

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
