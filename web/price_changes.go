package web

import (
	"net/http"
	"strconv"

	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/logger"
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
	changes, err := db.GetLatestPrices(priceChangeLimit, page)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Template
	t := priceChangesTemplate{}
	t.Fill(w, r, "Price Changes")
	t.Changes = changes
	t.Pagination = Pagination{
		path:  "/price-changes?p=",
		page:  page,
		limit: priceChangeLimit,
		total: priceChangeLimit * 10,
	}

	returnTemplate(w, r, "price_changes", t)
	return
}

type priceChangesTemplate struct {
	GlobalTemplate
	Apps       []db.App
	Changes    []db.Price
	Pagination Pagination
}
