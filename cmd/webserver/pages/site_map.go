package pages

import (
	"net/http"
	"time"

	"github.com/Jleagle/sitemap-go/sitemap"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

const urlBase = "https://gamedb.online"

func SiteMapIndexHandler(w http.ResponseWriter, r *http.Request) {

	var sitemaps = []string{
		"/sitemap-pages.xml",
		"/sitemap-games-by-score.xml",
		"/sitemap-games-by-players.xml",
		"/sitemap-players-by-level.xml",
		"/sitemap-players-by-games.xml",
		"/sitemap-groups.xml",
		"/sitemap-badges.xml",
	}

	sm := sitemap.NewSiteMapIndex()

	for _, v := range sitemaps {
		sm.AddSitemap(urlBase+v, time.Time{})
	}

	_, err := sm.Write(w)
	log.Err(err)
}

//noinspection GoUnusedParameter
func SiteMapPagesHandler(w http.ResponseWriter, r *http.Request) {

	var pages = []string{
		"/",
		"/api",
		"/apps/achievements",
		"/apps/random",
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
		"/login",
		"/news",
		"/packages",
		"/players",
		"/price-changes",
		"/publishers",
		"/stats",
		"/steam-api",
		"/tags",
		"/upcoming",
	}

	sm := sitemap.NewSitemap()

	for _, v := range pages {
		sm.AddLocation(urlBase+v, time.Time{}, sitemap.FrequencyHourly, 1)
	}

	_, err := sm.Write(w)
	log.Err(err)
}

func SiteMapGamesByScoreHandler(w http.ResponseWriter, r *http.Request) {

	apps, err := mongo.GetApps(0, 1000, bson.D{{"reviews_score", -1}}, bson.D{}, bson.M{"_id": 1, "name": 1, "updated_at": 1}, nil)
	if err != nil {
		log.Err(err, r)
		return
	}

	sm := sitemap.NewSitemap()
	for _, app := range apps {
		sm.AddLocation(urlBase+app.GetPath(), app.UpdatedAt, sitemap.FrequencyWeekly, 0.9)
	}

	_, err = sm.Write(w)
	if err != nil {
		log.Err(err, r)
	}
}

func SiteMapGamesByPlayersHandler(w http.ResponseWriter, r *http.Request) {

	apps, err := mongo.GetApps(0, 1000, bson.D{{"player_peak_week", -1}}, bson.D{}, bson.M{"_id": 1, "name": 1, "updated_at": 1}, nil)
	if err != nil {
		log.Err(err, r)
		return
	}

	sm := sitemap.NewSitemap()
	for _, app := range apps {
		sm.AddLocation(urlBase+app.GetPath(), app.UpdatedAt, sitemap.FrequencyWeekly, 0.9)
	}

	_, err = sm.Write(w)
	if err != nil {
		log.Err(err, r)
	}
}

//noinspection GoUnusedParameter
func SiteMapPlayersByLevel(w http.ResponseWriter, r *http.Request) {

	sm := sitemap.NewSitemap()

	players, err := mongo.GetPlayers(0, 1000, bson.D{{Key: "level", Value: -1}}, nil, bson.M{"_id": 1, "persona_name": 1, "updated_at": 1})
	log.Err(err)
	for _, player := range players {
		sm.AddLocation(urlBase+player.GetPath(), player.UpdatedAt, sitemap.FrequencyWeekly, 0.9)
	}

	_, err = sm.Write(w)
	log.Err(err)
}

//noinspection GoUnusedParameter
func SiteMapGroups(w http.ResponseWriter, r *http.Request) {

	sm := sitemap.NewSitemap()

	groups, err := mongo.GetGroups(1000, 0, bson.D{{Key: "members", Value: -1}}, bson.D{{Key: "type", Value: helpers.GroupTypeGroup}}, bson.M{"_id": 1, "name": 1, "updated_at": 1})
	log.Err(err)
	for _, v := range groups {
		sm.AddLocation(urlBase+v.GetPath(), v.UpdatedAt, sitemap.FrequencyWeekly, 0.9)
	}

	_, err = sm.Write(w)
	log.Err(err)
}

func SiteMapBadges(w http.ResponseWriter, r *http.Request) {

	sm := sitemap.NewSitemap()

	for _, badge := range mongo.GlobalBadges {
		sm.AddLocation(urlBase+badge.GetPath(), time.Time{}, sitemap.FrequencyWeekly, 0.9)
	}

	_, err := sm.Write(w)
	log.Err(err)
}

//noinspection GoUnusedParameter
func SiteMapPlayersByGamesCount(w http.ResponseWriter, r *http.Request) {

	sm := sitemap.NewSitemap()

	players, err := mongo.GetPlayers(0, 1000, bson.D{{Key: "games_count", Value: -1}}, nil, bson.M{"_id": 1, "persona_name": 1, "updated_at": 1})
	log.Err(err)
	for _, player := range players {
		sm.AddLocation(urlBase+player.GetPath(), player.UpdatedAt, sitemap.FrequencyWeekly, 0.9)
	}

	_, err = sm.Write(w)
	log.Err(err)
}
