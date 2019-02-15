package web

import (
	"encoding/xml"
	"net/http"

	"github.com/gamedb/website/config"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
	"github.com/go-chi/chi"
)

type ChangeFrequency string

//noinspection GoUnusedConst
const (
	urlBase   = "https://gamedb.online"
	namespace = "http://www.sitemaps.org/schemas/sitemap/0.9"

	frequencyAlways  ChangeFrequency = "always"
	frequencyHourly                  = "hourly"
	frequencyDaily                   = "daily"
	frequencyWeekly                  = "weekly"
	frequencyMonthly                 = "monthly"
	frequencyYearly                  = "yearly"
	frequencyNever                   = "never"
)

func siteMapRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", siteMapIndexHandler)
	r.Get("/pages", siteMapPagesHandler)
	r.Get("/games-by-score", siteMapGamesByScoreHandler)
	r.Get("/games-by-players", siteMapGamesByPlayersHandler)
	return r
}

type siteMapIndex struct {
	XMLName   xml.Name  `xml:"sitemapindex"`
	Namespace string    `xml:"xmlns,attr"`
	SiteMaps  []siteMap `xml:"sitemap"`
}

type siteMap struct {
	Location     string `xml:"loc"`
	LastModified string `xml:"lastmod,omitempty"` // https://www.w3.org/TR/NOTE-datetime
}

func siteMapIndexHandler(w http.ResponseWriter, r *http.Request) {

	sm := siteMapIndex{}
	sm.Namespace = namespace
	sm.SiteMaps = []siteMap{
		{Location: urlBase + "/sitemaps/pages"},
		{Location: urlBase + "/sitemaps/games-by-score"},
		{Location: urlBase + "/sitemaps/games-by-players"},
	}

	b, err := xml.Marshal(sm)
	log.Err(err)

	w.Header().Set("Content-Type", "application/xml")

	_, err = w.Write([]byte(xml.Header + string(b)))
	log.Err(err)
}

type urlSet struct {
	XMLName   xml.Name     `xml:"urlset"`
	Namespace string       `xml:"xmlns,attr"`
	URLs      []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Location        string          `xml:"loc"`
	LastModified    string          `xml:"lastmod,omitempty"` // https://www.w3.org/TR/NOTE-datetime
	ChangeFrequency ChangeFrequency `xml:"changefreq,omitempty"`
	Priority        float32         `xml:"priority,omitempty"`
}

func siteMapPagesHandler(w http.ResponseWriter, r *http.Request) {

	sm := urlSet{}
	sm.Namespace = namespace

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
		"/apps",
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

	for _, v := range pages {
		sm.URLs = append(sm.URLs, sitemapURL{
			Location: v,
			Priority: 1,
		})
	}

	b, err := xml.Marshal(sm)
	log.Err(err)

	w.Header().Set("Content-Type", "application/xml")

	_, err = w.Write([]byte(xml.Header + string(b)))
	log.Err(err)
}

func siteMapGamesByScoreHandler(w http.ResponseWriter, r *http.Request) {

	sm := urlSet{}
	sm.Namespace = namespace

	for _, v := range sitemapGetGames(r, "reviews_score desc") {

		sm.URLs = append(sm.URLs, sitemapURL{
			Location:        urlBase + v.GetPath(),
			Priority:        0.5,
			ChangeFrequency: frequencyWeekly,
		})
	}

	b, err := xml.Marshal(sm)
	log.Err(err)

	w.Header().Set("Content-Type", "application/xml")

	_, err = w.Write([]byte(xml.Header + string(b)))
	log.Err(err)
}

func siteMapGamesByPlayersHandler(w http.ResponseWriter, r *http.Request) {

	sm := urlSet{}
	sm.Namespace = namespace

	for _, v := range sitemapGetGames(r, "player_count desc") {

		sm.URLs = append(sm.URLs, sitemapURL{
			Location:        urlBase + v.GetPath(),
			Priority:        0.5,
			ChangeFrequency: frequencyWeekly,
		})
	}

	b, err := xml.Marshal(sm)
	log.Err(err)

	w.Header().Set("Content-Type", "application/xml")

	_, err = w.Write([]byte(xml.Header + string(b)))
	log.Err(err)
}

func sitemapGetGames(r *http.Request, sort string) (apps []db.App) {

	// Add most played apps
	gorm, err := db.GetMySQLClient()
	if err != nil {
		log.Err(err, r)
		return
	}

	gorm = gorm.Select([]string{"id", "name"})

	if config.Config.IsLocal() {
		gorm = gorm.Limit(10)
	} else {
		gorm = gorm.Limit(10000) // Max: 50,000
	}

	gorm = gorm.Order(sort)
	gorm = gorm.Find(&apps)

	log.Err(gorm.Error)
	return
}
