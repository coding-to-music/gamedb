package web

import (
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

const (
	FREE      = "free"
	CHANGES   = "changes"
	DISCOUNTS = "discounts"
)

func DealsHandler(w http.ResponseWriter, r *http.Request) {

	tab := chi.URLParam(r, "id")
	if tab == "" {
		tab = CHANGES
	}

	template := dealsTemplate{}
	template.Fill(r, "Deals")
	template.Tab = tab

	if tab == FREE {
		search := url.Values{}
		search.Set("is_free", "1")
		search.Set("name", "-")
		search.Set("type", "game")

		// Types not in this list will show first
		sort := "FIELD(`type`,'game','dlc','demo','mod','video','movie','series','episode','application','tool','advertising'), name ASC"
		freeApps, err := mysql.SearchApps(search, 1000, sort, []string{"id", "name", "icon", "type", "platforms"})
		if err != nil {
			logger.Error(err)
		}

		template.Apps = freeApps
	}

	if tab == CHANGES {
		changes, err := datastore.GetAppChanges()
		if err != nil {
			logger.Error(err)
		}

		template.Changes = changes
	}

	returnTemplate(w, r, "deals", template)
	return
}

type dealsTemplate struct {
	GlobalTemplate
	Apps    []mysql.App
	Tab     string
	Changes []datastore.AppPrice
}
