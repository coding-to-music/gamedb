package web

import (
	"net/http"
	"strconv"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

const (
	priceChangeLimit = 100
)

func PriceChangesHandler(w http.ResponseWriter, r *http.Request) {

	// Get page number
	page, err := strconv.Atoi(r.URL.Query().Get("p"))
	if err != nil {
		page = 1
	}

	// Get changes
	changes, err := datastore.GetLatestPrices(priceChangeLimit, page)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Template
	template := priceChangesTemplate{}
	template.Fill(w, r, "Price Changes")
	template.Changes = changes
	template.Pagination = Pagination{
		path:  "/price-changes?p=",
		page:  page,
		limit: priceChangeLimit,
		total: priceChangeLimit * 10,
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
