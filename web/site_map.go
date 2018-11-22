package web

import (
	"net/http"

	"github.com/gamedb/website/logging"
	"github.com/ikeikeikeike/go-sitemap-generator/stm"
)

var pages = []string{
	"/",
	"/changes",
	"/chat",
	"/commits",
	"/contact",
	"/coop",
	"/developers",
	"/discounts",
	"/donate",
	"/experience",
	"/free-games",
	"/games",
	"/genres",
	"/info",
	"/login",
	"/news",
	"/packages",
	"/players",
	"/price-changes",
	"/publishers",
	"/queues",
	"/stats",
	"/tags",
	"/upcoming",
}

func siteMapHandler(w http.ResponseWriter, r *http.Request) {

	sm := stm.NewSitemap(1)
	sm.SetDefaultHost("https://gamedb.online/")
	sm.SetCompress(true)
	sm.Create()

	for _, v := range pages {
		sm.Add(stm.URL{
			{"loc", v},
			{"changefreq", "daily"},
			{"mobile", true},
		})
	}

	_, err := w.Write(sm.XMLContent())
	logging.Error(err)
}
