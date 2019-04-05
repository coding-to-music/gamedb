package web

import (
	"net/http"
	"time"

	"github.com/Jleagle/sitemap-go/sitemap"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/mongo"
	"github.com/gamedb/website/sql"
	"github.com/go-chi/chi"
)

const urlBase = "https://gamedb.online"

func siteMapRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/pages.xml", siteMapPagesHandler)
	r.Get("/games-by-score.xml", siteMapGamesByScoreHandler)
	r.Get("/games-by-players.xml", siteMapGamesByPlayersHandler)
	r.Get("/players-by-level.xml", siteMapPlayersByLevel)
	r.Get("/players-by-games.xml", siteMapPlayersByGamesCount)
	return r
}

//noinspection GoUnusedParameter
func siteMapIndexHandler(w http.ResponseWriter, r *http.Request) {

	var sitemaps = []string{
		"/sitemap/pages.xml",
		"/sitemap/games-by-score.xml",
		"/sitemap/games-by-players.xml",
		"/sitemap/players-by-level.xml",
		"/sitemap/players-by-games.xml",
	}

	sm := sitemap.NewSiteMapIndex()

	for _, v := range sitemaps {
		sm.AddSitemap(urlBase+v, time.Time{})
	}

	_, err := sm.Write(w)
	log.Err(err)
}

//noinspection GoUnusedParameter
func siteMapPagesHandler(w http.ResponseWriter, r *http.Request) {

	var pages = []string{
		"/",
		"/changes",
		"/chat",
		"/commits",
		"/contact",
		"/coop",
		"/developers",
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
		"/steam-api",
		"/tags",
		"/upcoming",
	}

	sm := sitemap.NewSitemap()

	for _, v := range pages {
		sm.AddLocation(urlBase+v, time.Time{}, sitemap.FrequencyDaily, 1)
	}

	_, err := sm.Write(w)
	log.Err(err)
}

func siteMapGamesByScoreHandler(w http.ResponseWriter, r *http.Request) {

	sm := sitemap.NewSitemap()

	for _, v := range sitemapGetGames(r, "reviews_score desc") {
		sm.AddLocation(urlBase+v.GetPath(), time.Time{}, sitemap.FrequencyWeekly, 0.5)
	}

	_, err := sm.Write(w)
	log.Err(err)
}

func siteMapGamesByPlayersHandler(w http.ResponseWriter, r *http.Request) {

	sm := sitemap.NewSitemap()

	for _, v := range sitemapGetGames(r, "player_peak_week desc") {
		sm.AddLocation(urlBase+v.GetPath(), time.Time{}, sitemap.FrequencyWeekly, 0.5)
	}

	_, err := sm.Write(w)
	log.Err(err)
}

//noinspection GoUnusedParameter
func siteMapPlayersByLevel(w http.ResponseWriter, r *http.Request) {

	sm := sitemap.NewSitemap()

	players, err := mongo.GetPlayers(0, 1000, mongo.D{{"level", -1}}, nil, mongo.M{"_id": 1, "name": 1})
	for _, v := range players {
		sm.AddLocation(urlBase+v.GetPath(), time.Time{}, sitemap.FrequencyWeekly, 0.5)
	}

	_, err = sm.Write(w)
	log.Err(err)
}

//noinspection GoUnusedParameter
func siteMapPlayersByGamesCount(w http.ResponseWriter, r *http.Request) {

	sm := sitemap.NewSitemap()

	players, err := mongo.GetPlayers(0, 1000, mongo.D{{"games_count", -1}}, nil, mongo.M{"_id": 1, "name": 1})
	for _, v := range players {
		sm.AddLocation(urlBase+v.GetPath(), time.Time{}, sitemap.FrequencyWeekly, 0.5)
	}

	_, err = sm.Write(w)
	log.Err(err)
}

func sitemapGetGames(r *http.Request, sort string) (apps []sql.App) {

	// Add most played apps
	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err, r)
		return
	}

	gorm = gorm.Select([]string{"id", "name"})

	if config.Config.IsLocal() {
		gorm = gorm.Limit(10)
	} else {
		gorm = gorm.Limit(1000) // Max: 50,000
	}

	gorm = gorm.Order(sort)
	gorm = gorm.Find(&apps)

	log.Err(gorm.Error)
	return
}
