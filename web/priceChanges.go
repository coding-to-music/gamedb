package web

import (
	"net/http"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

func PriceChangesHandler(w http.ResponseWriter, r *http.Request) {

	changes, err := datastore.GetAppChanges()
	if err != nil {
		logger.Error(err)
	}

	template := priceChangesTemplate{}
	template.Fill(r, "Price Changes")
	template.Changes = changes

	returnTemplate(w, r, "price_changes", template)
	return
}

type priceChangesTemplate struct {
	GlobalTemplate
	Apps    []mysql.App
	Changes []datastore.AppPrice
}
