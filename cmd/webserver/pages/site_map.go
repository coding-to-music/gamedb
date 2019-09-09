package pages

import (
	"net/http"
	"time"

	"github.com/Jleagle/sitemap-go/sitemap"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
)

const urlBase = "https://gamedb.online"

func SiteMapRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/index.xml", siteMapIndexHandler)
	r.Get("/pages.xml", siteMapPagesHandler)
	r.Get("/games-by-score.xml", siteMapGamesByScoreHandler)
	r.Get("/games-by-players.xml", siteMapGamesByPlayersHandler)
	r.Get("/players-by-level.xml", siteMapPlayersByLevel)
	r.Get("/players-by-games.xml", siteMapPlayersByGamesCount)
	r.Get("/groups.xml", siteMapGroups)
	r.Get("/badges.xml", siteMapBadges)
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
		"/sitemap/groups.xml",
		"/sitemap/badges.xml",
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
		"/api",
		"/badges",
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
		"/groups",
		"/info",
		"/lp" + LandingAPI,
		"/lp" + LandingDeals,
		"/lp" + LandingTopGames,
		"/lp" + LandingXP,
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
		sm.AddLocation(urlBase+v, time.Time{}, sitemap.FrequencyMonthly, 1)
	}

	_, err := sm.Write(w)
	log.Err(err)
}

func siteMapGamesByScoreHandler(w http.ResponseWriter, r *http.Request) {

	sm := sitemap.NewSitemap()

	for _, app := range sitemapGetApps(r, "reviews_score desc") {
		sm.AddLocation(urlBase+app.GetPath(), app.UpdatedAt, sitemap.FrequencyWeekly, 0.9)
	}

	_, err := sm.Write(w)
	log.Err(err)
}

func siteMapGamesByPlayersHandler(w http.ResponseWriter, r *http.Request) {

	sm := sitemap.NewSitemap()

	for _, app := range sitemapGetApps(r, "player_peak_week desc") {
		sm.AddLocation(urlBase+app.GetPath(), app.UpdatedAt, sitemap.FrequencyWeekly, 0.9)
	}

	_, err := sm.Write(w)
	log.Err(err)
}

//noinspection GoUnusedParameter
func siteMapPlayersByLevel(w http.ResponseWriter, r *http.Request) {

	sm := sitemap.NewSitemap()

	players, err := mongo.GetPlayers(0, 1000, mongo.D{{"level", -1}}, nil, mongo.M{"_id": 1, "persona_name": 1, "updated_at": 1}, nil)
	for _, player := range players {
		sm.AddLocation(urlBase+player.GetPath(), player.UpdatedAt, sitemap.FrequencyWeekly, 0.9)
	}

	_, err = sm.Write(w)
	log.Err(err)
}

//noinspection GoUnusedParameter
func siteMapGroups(w http.ResponseWriter, r *http.Request) {

	sm := sitemap.NewSitemap()

	groups, err := mongo.GetGroups(1000, 0, mongo.D{{"members", -1}}, mongo.M{"type": "group"}, mongo.M{"_id": 1, "name": 1, "updated_at": 1})
	for _, v := range groups {
		sm.AddLocation(urlBase+v.GetPath(), v.UpdatedAt, sitemap.FrequencyWeekly, 0.9)
	}

	_, err = sm.Write(w)
	log.Err(err)
}

func siteMapBadges(w http.ResponseWriter, r *http.Request) {

	sm := sitemap.NewSitemap()

	for _, badge := range mongo.Badges {
		sm.AddLocation(urlBase+badge.GetPath(), time.Time{}, sitemap.FrequencyWeekly, 0.9)
	}

	_, err := sm.Write(w)
	log.Err(err)
}

//noinspection GoUnusedParameter
func siteMapPlayersByGamesCount(w http.ResponseWriter, r *http.Request) {

	sm := sitemap.NewSitemap()

	players, err := mongo.GetPlayers(0, 1000, mongo.D{{"games_count", -1}}, nil, mongo.M{"_id": 1, "persona_name": 1, "updated_at": 1}, nil)
	for _, player := range players {
		sm.AddLocation(urlBase+player.GetPath(), player.UpdatedAt, sitemap.FrequencyWeekly, 0.9)
	}

	_, err = sm.Write(w)
	log.Err(err)
}

func sitemapGetApps(r *http.Request, sort string) (apps []sql.App) {

	// Add most played apps
	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err, r)
		return
	}

	gorm = gorm.Select([]string{"id", "name", "updated_at"})
	gorm = gorm.Limit(1000) // Max: 50,000
	gorm = gorm.Order(sort)
	gorm = gorm.Find(&apps)

	log.Critical(gorm.Error)

	return apps
}
