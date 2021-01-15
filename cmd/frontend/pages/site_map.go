package pages

import (
	"net/http"
	"time"

	"github.com/Jleagle/sitemap-go/sitemap"
	"github.com/gamedb/gamedb/pkg/backend"
	"github.com/gamedb/gamedb/pkg/backend/generated"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

const siteMapLimit = 10_000

//noinspection GoUnusedParameter
func SiteMapIndexHandler(w http.ResponseWriter, r *http.Request) {

	var sitemaps = []string{
		"/sitemap-pages.xml",
		"/sitemap-games-by-score.xml",
		"/sitemap-games-by-players.xml",
		"/sitemap-games-new.xml",
		"/sitemap-games-upcoming.xml",
		"/sitemap-players-by-level.xml",
		"/sitemap-players-by-games.xml",
		"/sitemap-groups.xml",
		"/sitemap-badges.xml",
	}

	sm := sitemap.NewSiteMapIndex()

	for _, v := range sitemaps {
		sm.AddSitemap(config.C.GameDBDomain+v, time.Time{})
	}

	s, err := sm.String()
	if err != nil {
		log.ErrS(err)
	}

	returnXML(w, r, s)
}

//noinspection GoUnusedParameter
func SiteMapPagesHandler(w http.ResponseWriter, r *http.Request) {

	var pages = []string{
		"/",
		"/achievements",
		"/api",
		"/badges",
		"/changes",
		"/chat",
		"/commits",
		"/contact",
		"/developers",
		"/donate",
		"/experience",
		"/games",
		"/games/achievements",
		"/games/compare",
		"/games/coop",
		"/games/new-releases",
		"/games/random",
		"/games/sales",
		"/games/trending",
		"/games/upcoming",
		"/games/wishlist",
		"/games/release-dates",
		"/genres",
		"/groups",
		"/info",
		"/login",
		"/news",
		"/packages",
		"/players",
		"/price-changes",
		"/product-keys",
		"/publishers",
		"/stats",
		"/steam-api",
		"/tags",
	}

	sm := sitemap.NewSitemap()

	for _, page := range pages {
		sm.AddLocation(config.C.GameDBDomain+page, time.Time{}, sitemap.FrequencyHourly, 1)
	}

	s, err := sm.String()
	if err != nil {
		log.ErrS(err)
	}

	returnXML(w, r, s)
}

func SiteMapGamesByPlayersHandler(w http.ResponseWriter, r *http.Request) {

	apps, err := mongo.GetApps(0, siteMapLimit, bson.D{{"player_peak_week", -1}}, bson.D{}, bson.M{"_id": 1, "name": 1, "updated_at": 1})
	if err != nil {
		log.ErrS(err)
		return
	}

	sm := sitemap.NewSitemap()
	for _, app := range apps {
		sm.AddLocation(config.C.GameDBDomain+app.GetPath(), app.UpdatedAt, sitemap.FrequencyWeekly, 0.9)
	}

	s, err := sm.String()
	if err != nil {
		log.ErrS(err)
	}

	returnXML(w, r, s)
}

func SiteMapGamesByScoreHandler(w http.ResponseWriter, r *http.Request) {

	apps, err := mongo.GetApps(0, siteMapLimit, bson.D{{"reviews_score", -1}}, bson.D{}, bson.M{"_id": 1, "name": 1, "updated_at": 1})
	if err != nil {
		log.ErrS(err)
		return
	}

	sm := sitemap.NewSitemap()
	for _, app := range apps {
		sm.AddLocation(config.C.GameDBDomain+app.GetPath(), app.UpdatedAt, sitemap.FrequencyWeekly, 0.9)
	}

	s, err := sm.String()
	if err != nil {
		log.ErrS(err)
	}

	returnXML(w, r, s)
}

func SiteMapGamesUpcomingHandler(w http.ResponseWriter, r *http.Request) {

	apps, err := mongo.GetApps(0, siteMapLimit, bson.D{{"release_date_unix", 1}}, upcomingFilter, bson.M{"_id": 1, "name": 1, "updated_at": 1})
	if err != nil {
		log.ErrS(err)
		return
	}

	sm := sitemap.NewSitemap()
	for _, app := range apps {
		sm.AddLocation(config.C.GameDBDomain+app.GetPath(), app.UpdatedAt, sitemap.FrequencyWeekly, 0.9)
	}

	s, err := sm.String()
	if err != nil {
		log.ErrS(err)
	}

	returnXML(w, r, s)
}

func SiteMapGamesNewHandler(w http.ResponseWriter, r *http.Request) {

	var filter = bson.D{
		{"release_date_unix", bson.M{"$lt": time.Now().Unix()}},
		{"release_date_unix", bson.M{"$gt": time.Now().AddDate(0, 0, -config.C.NewReleaseDays).Unix()}},
	}

	apps, err := mongo.GetApps(0, siteMapLimit, bson.D{{"release_date_unix", -1}}, filter, bson.M{"_id": 1, "name": 1, "updated_at": 1})
	if err != nil {
		log.ErrS(err)
		return
	}

	sm := sitemap.NewSitemap()
	for _, app := range apps {
		sm.AddLocation(config.C.GameDBDomain+app.GetPath(), app.UpdatedAt, sitemap.FrequencyWeekly, 0.9)
	}

	s, err := sm.String()
	if err != nil {
		log.ErrS(err)
	}

	returnXML(w, r, s)
}

//noinspection GoUnusedParameter
func SiteMapPlayersByLevel(w http.ResponseWriter, r *http.Request) {

	sm := sitemap.NewSitemap()

	players, err := mongo.GetPlayers(0, siteMapLimit, bson.D{{Key: "level", Value: -1}}, nil, bson.M{"_id": 1, "persona_name": 1, "updated_at": 1})
	if err != nil {
		log.ErrS(err)
	}

	for _, player := range players {
		sm.AddLocation(config.C.GameDBDomain+player.GetPath(), player.UpdatedAt, sitemap.FrequencyWeekly, 0.9)
	}

	s, err := sm.String()
	if err != nil {
		log.ErrS(err)
	}

	returnXML(w, r, s)
}

//noinspection GoUnusedParameter
func SiteMapPlayersByGamesCount(w http.ResponseWriter, r *http.Request) {

	sm := sitemap.NewSitemap()

	players, err := mongo.GetPlayers(0, siteMapLimit, bson.D{{Key: "games_count", Value: -1}}, nil, bson.M{"_id": 1, "persona_name": 1, "updated_at": 1})
	if err != nil {
		log.ErrS(err)
	}

	for _, player := range players {
		sm.AddLocation(config.C.GameDBDomain+player.GetPath(), player.UpdatedAt, sitemap.FrequencyWeekly, 0.9)
	}

	s, err := sm.String()
	if err != nil {
		log.ErrS(err)
	}

	returnXML(w, r, s)
}

//noinspection GoUnusedParameter
func SiteMapGroups(w http.ResponseWriter, r *http.Request) {

	conn, ctx, err := backend.GetClient()
	if err != nil {
		log.Err(err.Error())
		return
	}

	message := &generated.GroupsRequest{
		Pagination: &generated.PaginationRequest{
			Limit:     siteMapLimit,
			SortField: "members",
			SortOrder: "desc",
		},
		Type: helpers.GroupTypeGroup,
	}

	resp, err := generated.NewGroupsServiceClient(conn).List(ctx, message)
	if err != nil {
		log.Err(err.Error())
		return
	}

	sm := sitemap.NewSitemap()
	for _, group := range resp.GetGroups() {

		p := helpers.GetGroupPath(group.GetID(), group.GetName())

		sm.AddLocation(config.C.GameDBDomain+p, group.GetUpdatedAt().AsTime(), sitemap.FrequencyWeekly, 0.9)
	}

	s, err := sm.String()
	if err != nil {
		log.ErrS(err)
	}

	returnXML(w, r, s)
}

func SiteMapBadges(w http.ResponseWriter, r *http.Request) {

	sm := sitemap.NewSitemap()

	for _, badge := range helpers.BuiltInSpecialBadges {
		sm.AddLocation(config.C.GameDBDomain+badge.GetPath(false), time.Time{}, sitemap.FrequencyWeekly, 0.9)
	}
	for _, badge := range helpers.BuiltInEventBadges {
		sm.AddLocation(config.C.GameDBDomain+badge.GetPath(false), time.Time{}, sitemap.FrequencyWeekly, 0.9)
	}

	s, err := sm.String()
	if err != nil {
		log.ErrS(err)
	}

	returnXML(w, r, s)
}
