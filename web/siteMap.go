package web

import (
	"net/http"

	"github.com/ikeikeikeike/go-sitemap-generator/stm"
)

func SiteMapHandler(w http.ResponseWriter, r *http.Request) {

	sm := stm.NewSitemap()
	sm.SetDefaultHost("https://steamauthority.net/")
	sm.SetCompress(true)
	sm.Create()

	pages := []string{
		"/",
		"/changes",
		"/chat",
		"/commits",
		"/contact",
		"/coop",
		"/discounts",
		"/developers",
		"/donate",
		"/experience",
		"/free-games",
		"/games",
		"/genres",
		"/info",
		"/news",
		"/packages",
		"/players",
		"/price-changes",
		"/publishers",
		"/queues",
		"/stats",
		"/tags",
	}

	for _, v := range pages {
		sm.Add(stm.URL{"loc": v, "changefreq": "daily", "mobile": true})
	}

	w.Write(sm.XMLContent())
	return
}
