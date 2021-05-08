package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql/pics"
	"github.com/gamedb/gamedb/pkg/session"
	"github.com/go-chi/chi/v5"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

func appRouter() http.Handler {

	r := chi.NewRouter()
	r.Mount("/compare-achievements", appCompareAchievementsRouter())

	r.Get("/", appHandler)
	r.Get("/achievement-counts.json", appAchievementCountsAjaxHandler)
	r.Get("/achievements.json", appAchievementsAjaxHandler)
	r.Get("/bundles.json", appBundlesAjaxHandler)
	r.Get("/dlc.json", appDLCAjaxHandler)
	r.Get("/friends.json", friendsJSONHandler)
	r.Get("/items.json", appItemsAjaxHandler)
	r.Get("/localization.html", appLocalizationHandler)
	r.Get("/news.json", appNewsAjaxHandler)
	r.Get("/news-feeds.json", appNewsFeedsAjaxHandler)
	r.Get("/packages.json", appPackagesAjaxHandler)
	r.Get("/players-heatmap.json", appPlayersHeatmapAjaxHandler)
	r.Get("/players.json", appPlayersAjaxHandler(true))
	r.Get("/players2.json", appPlayersAjaxHandler(false))
	r.Get("/prices.json", appPricesAjaxHandler)
	r.Get("/reviews.html", appReviewsHandler)
	r.Get("/reviews.json", appReviewsAjaxHandler)
	r.Get("/same-owners.json", appSamdOwnersHandler)
	r.Get("/similar.html", appSimilarHandler)
	r.Get("/tags.json", appTagsAjaxHandler)
	r.Get("/time.json", appTimeAjaxHandler)
	r.Get("/wishlist.json", appWishlistAjaxHandler)
	r.Get("/{slug}", appHandler)

	return r
}

func appHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid App ID"})
		return
	}

	if !helpers.IsValidAppID(id) {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid App ID"})
		return
	}

	// Get app
	app, err := mongo.GetApp(id)
	if err != nil && strings.HasPrefix(err.Error(), "memcache: unexpected response line from \"set\":") {
		log.WarnS(err)
		err = nil
	}
	if err != nil {

		if err == mongo.ErrNoDocuments {
			returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Sorry but we can not find this app."})
			return
		}

		err = helpers.IgnoreErrors(err, mongo.ErrInvalidAppID)
		if err != nil {
			log.ErrS(err)
		}

		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the app."})
		return
	}

	// Template
	t := appTemplate{}
	t.setBackground(app, false, false)
	t.fill(w, r, "app", app.GetName(), template.HTML(app.ShortDescription))
	t.addAssetHighCharts()
	t.addAssetHighChartsHeatmap()
	t.addAssetJSON2HTML()
	t.addAssetMomentData()
	t.addAssetChosen()
	t.metaImage = app.GetMetaImage()
	t.App = app
	t.Description = template.HTML(app.ShortDescription)
	t.Canonical = app.GetPath()
	t.GroupPath = helpers.GetGroupPath(app.GroupID, app.GetName())

	for _, tagID := range app.Tags {
		if tagID == 24904 {
			t.hideAds = true
			break
		}
	}

	//
	var wg sync.WaitGroup

	// Update news, reviews etc
	wg.Add(1)
	go func() {

		defer wg.Done()

		if helpers.IsBot(r.UserAgent()) {
			return
		}

		if !app.ShouldUpdate() {
			return
		}

		err = consumers.ProduceSteam(consumers.SteamMessage{AppIDs: []int{app.ID}})
		if err == nil {
			t.addToast(Toast{Title: "Update", Message: "App has been queued for an update", Success: true})
			log.Info("app queued", zap.String("ua", r.UserAgent()))
		}
		err = helpers.IgnoreErrors(err, consumers.ErrInQueue)
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Tags
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Tags, err = app.GetTags()
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Categories
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Categories, err = app.GetCategories()
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Genres
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Genres, err = app.GetGenres()
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get demos
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Demos, err = app.GetDemos()
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get Developers
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Developers, err = app.GetDevelopers()
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get Publishers
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Publishers, err = app.GetPublishers()
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get players count
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PlayersCount, err = mongo.CountDocuments(mongo.CollectionPlayers, nil, 0)
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get played time
	wg.Add(1)
	var playedMessage string
	go func() {

		defer wg.Done()

		if session.IsLoggedIn(r) {

			playerID := session.GetPlayerIDFromSesion(r)
			if playerID > 0 {

				playerApp, err := mongo.GetPlayerAppByKey(playerID, app.ID)
				if err != nil {
					err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
					if err != nil {
						log.ErrS(err)
					}
					return
				}

				if playerApp.AppTime > 0 {
					playedMessage = "You own this game and have played for " + playerApp.GetTimeNice()
				} else {
					playedMessage = "You own this game, but have never played"
				}
			}
		}
	}()

	// Wait
	wg.Wait()

	// Countries
	delete(app.Countries, "")

	var countryTotal int
	for k, v := range app.Countries {
		countryTotal += v
		t.Countries = append(t.Countries, AppCountry{
			Country: i18n.CountryCodeToName(k),
			Count:   v,
		})
	}
	for k, v := range t.Countries {
		t.Countries[k].Percent = float64(v.Count) / float64(countryTotal) * 100
	}
	sort.Slice(t.Countries, func(i, j int) bool {
		return t.Countries[i].Count > t.Countries[j].Count
	})

	//
	t.PlayersInGame, err = t.App.GetPlayersInGame()
	if err != nil {
		log.ErrS(err)
	}

	// Functions that get called multiple times in the template
	t.Price = app.Prices.Get(session.GetProductCC(r))
	t.Common = app.Common.Formatted(app.ID, pics.CommonKeys)
	t.Extended = app.Extended.Formatted(app.ID, pics.ExtendedKeys)
	t.Config = app.Config.Formatted(app.ID, pics.ConfigKeys)
	t.UFS = app.UFS.Formatted(app.ID, pics.UFSKeys)

	// Sort reviews
	sort.Slice(app.Reviews.Reviews, func(i, j int) bool {
		return app.Reviews.Reviews[i].VotesGood > app.Reviews.Reviews[j].VotesGood
	})

	// Sort launch
	sort.Slice(app.Launch, func(i, j int) bool {
		return app.Launch[i].OrderInt() < app.Launch[j].OrderInt()
	})

	// Get max tag count
	t.TagsMax = 1
	for _, v := range app.TagCounts {
		if v.Count > t.TagsMax {
			t.TagsMax = v.Count
		}
	}

	// Make banners
	var banners = map[string][]string{
		"primary": {},
		"warning": {},
	}

	if app.ID == 753 {
		banners["primary"] = append(banners["primary"], "This app record is for the Steam client")
	}
	if playedMessage != "" {
		banners["primary"] = append(banners["primary"], playedMessage)
	}

	if app.ReadPICS(app.Common).GetValue("app_retired_publisher_request") == "1" {
		banners["warning"] = append(banners["warning"], "At the request of the publisher, "+app.GetName()+" is no longer available for sale on Steam.")
	}

	t.Banners = banners

	// Links
	t.Links = []appLinkTemplate{
		{
			Text: "View " + app.GetTypeLower() + " on Twitch",
			Link: "https://twitch.tv/directory/game/" + url.PathEscape(app.TwitchURL),
			Icon: "/assets/img/links/twitch.png",
			Hide: app.TwitchURL == "",
		},
		{
			Text: "View " + app.GetTypeLower() + " on Achievement Stats",
			Link: "https://www.achievementstats.com/index.php?action=games&gameId=" + strconv.Itoa(app.ID),
			Icon: "/assets/img/links/achievement-stats.ico",
		},
		{
			Text: "View " + app.GetTypeLower() + " on Steam Hunters",
			Link: "https://steamhunters.com/stats/" + strconv.Itoa(app.ID) + "/achievements",
			Icon: "/assets/img/links/steam-hunters.png",
		},
		{
			Text: "View " + app.GetTypeLower() + " on IsThereAnyDeal",
			Link: "https://isthereanydeal.com/steam/app/" + strconv.Itoa(app.ID),
			Icon: "/assets/img/links/is-there-any-deal.png",
		},
		{
			Text: "Find similar " + app.GetTypeLower() + "s on SteamPeek",
			Link: "https://steampeek.hu/?appid=" + strconv.Itoa(app.ID),
			Icon: "/assets/img/links/steam-peek.png",
		},
	}

	// Get user
	t.APIKey = session.Get(r, session.SessionUserAPIKey)

	// Heatmap Timezones
	for i := -12; i <= 12; i++ {
		s := strconv.Itoa(i)
		if i >= 0 {
			s = "+" + s
		}
		t.Timezones = append(t.Timezones, s)
	}

	//
	returnTemplate(w, r, t)
}

type appTemplate struct {
	globalTemplate
	App           mongo.App
	PlayersCount  int64
	Banners       map[string][]string
	Common        []pics.KeyValue
	Config        []pics.KeyValue
	Demos         []mongo.App
	Extended      []pics.KeyValue
	Links         []appLinkTemplate
	Price         helpers.ProductPrice
	TagsMax       int
	UFS           []pics.KeyValue
	PlayersInGame int64
	GroupPath     string
	Countries     []AppCountry
	APIKey        string
	Timezones     []string

	// Stats
	Categories []mongo.Stat
	Developers []mongo.Stat
	Genres     []mongo.Stat
	Publishers []mongo.Stat
	Tags       []mongo.Stat
}

type AppCountry struct {
	Country string
	Count   int
	Percent float64
}

func (ac AppCountry) GetPercent() string {
	return helpers.FloatToString(ac.Percent, 2)
}

type appLinkTemplate struct {
	Text string
	Link string
	Icon string
	Hide bool
}

func appLocalizationHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return
	}

	if !helpers.IsValidAppID(id) {
		return
	}

	app := mongo.App{}
	err = mongo.FindOne(mongo.CollectionApps, bson.D{{Key: "_id", Value: id}}, nil, bson.M{"localization": 1}, &app)
	if err != nil {
		err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
		if err != nil {
			log.ErrS(err)
		}
		return
	}

	t := appLocalizationTemplate{}
	t.fill(w, r, "app_localization", "", "")
	t.App = app

	returnTemplate(w, r, t)
}

type appLocalizationTemplate struct {
	globalTemplate
	App mongo.App
}

func appSamdOwnersHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return
	}

	if !helpers.IsValidAppID(id) {
		return
	}

	var limit int64 = 100
	if r.URL.Query().Get("small") == "1" {
		limit = 5
	}

	sameApps, err := mongo.GetAppSameOwners(id, limit)
	if err != nil {
		log.ErrS(err)
		return
	}

	var countsMap = map[int]mongo.AppSameOwners{}
	var appIDs []int
	for _, sameApp := range sameApps {
		appIDs = append(appIDs, sameApp.SameAppID)
		countsMap[sameApp.SameAppID] = sameApp
	}

	apps, err := mongo.GetAppsByID(appIDs, bson.M{"_id": 1, "name": 1, "icon": 1})
	if err != nil {
		log.ErrS(err)
		return
	}

	appsMap := map[int]mongo.App{}
	for _, v := range apps {
		appsMap[v.ID] = v
	}

	var orderedApps []mongo.App
	for _, v := range sameApps {
		orderedApps = append(orderedApps, appsMap[v.SameAppID])
	}

	//
	var query = datatable.NewDataTableQuery(r, false)
	var response = datatable.NewDataTablesResponse(r, query, int64(len(orderedApps)), int64(len(orderedApps)), nil)
	for _, app := range orderedApps {

		response.AddRow([]interface{}{
			app.ID,                     // 0
			app.GetPath() + "#similar", // 1
			app.GetIcon(),              // 2
			app.GetName(),              // 3
			app.GetStoreLink(),         // 4
			countsMap[app.ID].Count,    // 5
			countsMap[app.ID].Owners,   // 6
			countsMap[app.ID].Order,    // 7
		})
	}

	returnJSON(w, r, response)
}

func appSimilarHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return
	}

	if !helpers.IsValidAppID(id) {
		return
	}

	app, err := mongo.GetApp(id)
	if err != nil {
		log.ErrS(err)
		return
	}

	related, err := app.GetAppRelatedApps()
	if err != nil {
		log.ErrS(err)
		return
	}

	var tagIDs []int
	for _, v := range related {
		for _, vv := range v.Tags {
			if helpers.SliceHasInt(app.Tags, vv) {
				tagIDs = append(tagIDs, vv)
			}
		}
	}

	tags, err := mongo.GetStatsByID(mongo.StatsTypeTags, tagIDs)
	if err != nil {
		log.ErrS(err)
		return
	}

	relatedTags := map[int]mongo.Stat{}
	for _, v := range tags {
		relatedTags[v.ID] = v
	}

	t := appSimilarTemplate{}
	t.fill(w, r, "app_similar", "", "")
	t.Related = related
	t.RelatedTags = relatedTags

	returnTemplate(w, r, t)
}

type appSimilarTemplate struct {
	globalTemplate
	RelatedTags map[int]mongo.Stat
	Related     []mongo.App
}

func (t appSimilarTemplate) GetRelatedTags(relatedApp mongo.App) template.HTML {

	var ret []string
	for _, v := range relatedApp.Tags {
		if val, ok := t.RelatedTags[v]; ok {
			ret = append(ret, `<a href="`+val.GetPath()+`">`+val.Name+`</a>`)
		}
	}

	return template.HTML(strings.Join(ret, ", "))
}

func appReviewsHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return
	}

	if !helpers.IsValidAppID(id) {
		return
	}

	app := mongo.App{}
	err = mongo.FindOne(mongo.CollectionApps, bson.D{{Key: "_id", Value: id}}, nil, bson.M{"reviews": 1}, &app)
	if err != nil {
		err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
		if err != nil {
			log.ErrS(err)
		}
		return
	}

	t := appReviewsTemplate{}
	t.fill(w, r, "app_reviews", "", "")
	t.App = app

	returnTemplate(w, r, t)
}

type appReviewsTemplate struct {
	globalTemplate
	App mongo.App
}

func appNewsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid App ID"})
		return
	}

	var query = datatable.NewDataTableQuery(r, false)
	var community = r.URL.Query().Get("community") != ""
	var wg sync.WaitGroup

	// Get articles
	var articles []elasticsearch.Article
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var filters = []elastic.Query{elastic.NewTermQuery("app_id", id)}
		var sorter = []elastic.Sorter{elastic.NewFieldSort("time").Desc()}
		var limit int

		// Feed
		var feeds = query.GetSearchSliceInterface("article-feed")
		if len(feeds) > 0 {
			filters = append(filters, elastic.NewTermsQuery("feed", feeds...))
		}
		if community {
			filters = append(filters, elastic.NewTermsQuery("feed", "steam_community_announcements"))
			limit = 5
		}

		var err error
		articles, filtered, err = elasticsearch.SearchArticles(query.GetOffset(), limit, sorter, query.GetSearchString("search"), filters)
		if err != nil {
			log.ErrS(err, id)
			return
		}
	}()

	// Get totals
	var total int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		if community {
			return
		}

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionAppArticles, bson.D{{Key: "app_id", Value: id}}, 60*60*24)
		if err != nil {
			log.ErrS(err)
			return
		}
	}()

	wg.Wait()

	if community {
		filtered = 5
	}

	//
	var response = datatable.NewDataTablesResponse(r, query, total, filtered, nil)
	for _, article := range articles {

		var id = strconv.FormatInt(article.ID, 10)
		var path = helpers.GetAppPath(article.AppID, article.AppName)
		var formatted = time.Unix(article.Time, 0).Format(helpers.DateYear)

		response.AddRow([]interface{}{
			id,                       // 0
			article.Title,            // 1
			article.Author,           // 2
			article.Time,             // 3
			formatted,                // 4
			article.GetBody(),        // 5
			article.AppID,            // 6
			article.AppName,          // 7
			article.GetAppIcon(),     // 8
			path + "#news," + id,     // 9
			article.GetArticleIcon(), // 10
		})
	}

	returnJSON(w, r, response)
}

func appNewsFeedsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid App ID"})
		return
	}

	feeds, err := mongo.GetAppArticlesGroupedByFeed(id)
	if err != nil {
		log.ErrS(err)
	}

	returnJSON(w, r, feeds)
}

func appPricesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	productPricesAjaxHandler(w, r, helpers.ProductTypeApp)
}

func appAchievementCountsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return
	}

	counts, err := mongo.GetAchievmentCounts(id)
	if err != nil {
		log.ErrS(err)
		return
	}

	var marker int
	playerID := session.GetPlayerIDFromSesion(r)
	if playerID > 0 {
		playerApp, err := mongo.GetPlayerAppByKey(playerID, id)
		if err != nil {
			err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
			if err != nil {
				log.ErrS(err)
			}
		} else {
			marker = playerApp.AppAchievementsHave
		}
	}

	var ret [][]int
	for _, v := range counts {
		ret = append(ret, []int{v.ID, v.Count})
	}

	returnJSON(w, r, map[string]interface{}{
		"data":   ret,
		"marker": marker,
	})
}

func appAchievementsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return
	}

	// Get player achievements
	var achievedMap = map[string]int64{}
	var filter = bson.D{{Key: "app_id", Value: id}}
	var query = datatable.NewDataTableQuery(r, false)

	var playerID = session.GetPlayerIDFromSesion(r)
	if playerID > 0 {

		playerAchievements, err := mongo.GetPlayerAchievementsByPlayerAndApp(playerID, id)
		if err != nil {
			log.ErrS(err)
			return
		}

		var achievementKeys = bson.A{"_"} // $in & $nin can't have empty arrays
		for _, v := range playerAchievements {
			achievedMap[v.AchievementID] = v.AchievementDate
			achievementKeys = append(achievementKeys, v.AchievementID)
		}

		var search = query.GetSearchString("achievements-filter")
		if search == "locked" {
			filter = append(filter, bson.E{Key: "key", Value: bson.M{"$nin": achievementKeys}})
		} else if search == "unlocked" {
			filter = append(filter, bson.E{Key: "key", Value: bson.M{"$in": achievementKeys}})
		}
	}

	//
	var wg sync.WaitGroup

	wg.Add(1)
	var appAchievements []mongo.AppAchievement
	go func() {

		defer wg.Done()

		var err error
		var sortOrder = query.GetOrderMongo(map[string]string{
			"2": "completed",
		})

		appAchievements, err = mongo.GetAppAchievements(query.GetOffset64(), 1000, filter, sortOrder)
		if err != nil {
			log.ErrS(err)
			return
		}
	}()

	wg.Add(1)
	var total int64
	go func() {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionAppAchievements, filter, 60*60*24*28)
		if err != nil {
			log.ErrS(err)
			return
		}
	}()

	wg.Wait()

	//
	response := datatable.NewDataTablesResponse(r, query, total, total, nil)
	for _, appCchievement := range appAchievements {

		achievedTime, ok := achievedMap[appCchievement.Key]
		if ok && achievedTime == 0 {
			achievedTime = -1
		}

		var achievedTimeFormatted string
		if achievedTime != 0 {
			achievedTimeFormatted = time.Unix(achievedTime, 0).Format(helpers.DateSQLMinute)
		}

		response.AddRow([]interface{}{
			appCchievement.Name,           // 0
			appCchievement.Description,    // 1
			appCchievement.GetIcon(),      // 2
			appCchievement.GetCompleted(), // 3
			appCchievement.Active,         // 4
			appCchievement.Hidden,         // 5
			appCchievement.Deleted,        // 6
			achievedTime,                  // 7
			achievedTimeFormatted,         // 8
			appCchievement.Key,            // 9
		})
	}

	returnJSON(w, r, response)
}

//
func appDLCAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return
	}

	var query = datatable.NewDataTableQuery(r, false)
	var search = query.GetSearchString("search")

	var filter = bson.D{{Key: "app_id", Value: id}}
	var filter2 = filter

	if len(search) > 1 {

		quoted := regexp.QuoteMeta(search)

		filter2 = append(filter2, bson.E{Key: "name", Value: bson.M{"$regex": quoted, "$options": "i"}})
	}

	//
	var wg sync.WaitGroup

	// Get DLCs
	var dlcs []mongo.AppDLC
	wg.Add(1)
	go func() {

		defer wg.Done()

		var sortOrder = query.GetOrderMongo(map[string]string{
			"0": "name",
			"1": "release_date_unix",
		})

		var err error
		dlcs, err = mongo.GetDLCForApp(query.GetOffset64(), 100, filter2, sortOrder, nil)
		if err != nil {
			log.ErrS(err)
			return
		}
	}()

	// Get totals
	var total int64
	var filtered int64

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		total, err = mongo.CountDocuments(mongo.CollectionAppDLC, filter, 60*60*24)
		if err != nil {
			log.ErrS(err)
			return
		}

		filtered, err = mongo.CountDocuments(mongo.CollectionAppDLC, filter2, 60*60*24)
		if err != nil {
			log.ErrS(err)
			return
		}
	}()

	// Wait
	wg.Wait()

	response := datatable.NewDataTablesResponse(r, query, total, filtered, nil)
	for _, dlc := range dlcs {

		response.AddRow([]interface{}{
			dlc.DLCID,           // 0
			dlc.GetName(),       // 1
			dlc.GetIcon(),       // 2
			dlc.ReleaseDateUnix, // 3
			dlc.ReleaseDateNice, // 4
			dlc.GetPath(),       // 5
		})
	}

	returnJSON(w, r, response)
}

func appItemsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.ErrS(err)
		return
	}

	var query = datatable.NewDataTableQuery(r, false)
	var search = query.GetSearchString("search")
	var filter = bson.D{{Key: "app_id", Value: id}}
	var filter2 = filter

	if len(search) > 1 {

		quoted := regexp.QuoteMeta(search)

		filter2 = append(filter2, bson.E{Key: "$or", Value: bson.A{
			bson.M{"name": bson.M{"$regex": quoted, "$options": "i"}},
			bson.M{"description": bson.M{"$regex": quoted, "$options": "i"}},
		}})
	}

	//
	var wg sync.WaitGroup

	// Get items
	var items []mongo.AppItem
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		items, err = mongo.GetAppItems(query.GetOffset64(), 100, filter2, nil)
		if err != nil {
			log.ErrS(err)
			return
		}

	}()

	// Get totals
	var total int64
	var filtered int64

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		total, err = mongo.CountDocuments(mongo.CollectionAppItems, filter, 0)
		if err != nil {
			log.ErrS(err)
		}

		filtered, err = mongo.CountDocuments(mongo.CollectionAppItems, filter2, 0)
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, total, filtered, nil)
	for _, item := range items {

		var image1 = item.Image(54, true)
		var image2 = item.Image(256, false)

		response.AddRow([]interface{}{
			item.AppID,              // 0
			item.Bundle,             // 1
			item.Commodity,          // 2
			item.DateCreated,        // 3
			item.Description,        // 4
			item.DisplayType,        // 5
			item.DropInterval,       // 6
			item.DropMaxPerWindow,   // 7
			item.Exchange,           // 8
			item.Hash,               // 9
			item.IconURL,            // 10
			item.IconURLLarge,       // 11
			item.ItemDefID,          // 12
			item.ItemQuality,        // 13
			item.Marketable,         // 14
			item.Modified,           // 15
			item.Name,               // 16
			item.Price,              // 17
			item.Promo,              // 18
			item.Quantity,           // 19
			item.Tags,               // 20
			item.Timestamp,          // 21
			item.Tradable,           // 22
			item.Type,               // 23
			item.WorkshopID,         // 24
			image1,                  // 25
			image2,                  // 26
			item.GetType(),          // 27
			item.Link(),             // 28
			item.ShortDescription(), // 29
		})
	}

	returnJSON(w, r, response)
}

func appBundlesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(helpers.RegexIntsOnly.FindString(chi.URLParam(r, "id")))
	if err != nil || !helpers.IsValidAppID(id) {
		return
	}

	app, err := mongo.GetApp(id)
	if err != nil {
		log.ErrS(err)
		return
	}
	if len(app.Bundles) == 0 {
		return
	}

	var bundles []mongo.Bundle
	var callback = func() (interface{}, error) {
		return mongo.GetBundlesByID(app.Bundles, nil)
	}

	item := memcache.ItemAppBundles(app.ID)
	err = memcache.Client().GetSet(item.Key, item.Expiration, &bundles, callback)
	if err != nil {
		log.ErrS(err)
		return
	}

	var query = datatable.NewDataTableQuery(r, false)
	var response = datatable.NewDataTablesResponse(r, query, int64(len(app.Bundles)), int64(len(app.Bundles)), nil)
	for _, bundle := range bundles {
		response.AddRow([]interface{}{
			bundle.ID,               // 0
			bundle.GetPath(),        // 1
			bundle.GetName(),        // 2
			bundle.Discount,         // 3
			bundle.GetApps(),        // 4
			bundle.GetPackages(),    // 5
			bundle.GetUpdatedNice(), // 6
		})
	}

	returnJSON(w, r, response)
}

func appPackagesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(helpers.RegexIntsOnly.FindString(chi.URLParam(r, "id")))
	if err != nil || !helpers.IsValidAppID(id) {
		return
	}

	app, err := mongo.GetApp(id)
	if err != nil {
		log.ErrS(err)
		return
	}
	if len(app.Packages) == 0 {
		return
	}

	var packages []mongo.Package
	var callback = func() (interface{}, error) {
		return mongo.GetPackagesByID(app.Packages, bson.M{})
	}

	item := memcache.ItemAppPackages(app.ID)
	err = memcache.Client().GetSet(item.Key, item.Expiration, &packages, callback)
	if err != nil {
		log.ErrS(err)
		return
	}

	var query = datatable.NewDataTableQuery(r, false)
	var response = datatable.NewDataTablesResponse(r, query, int64(len(app.Packages)), int64(len(app.Packages)), nil)
	for _, pack := range packages {
		response.AddRow([]interface{}{
			pack.ID,               // 0
			pack.GetPath(),        // 1
			pack.GetName(),        // 2
			pack.GetBillingType(), // 3
			pack.GetLicenseType(), // 4
			pack.GetStatus(),      // 5
			pack.AppsCount,        // 6
		})
	}

	returnJSON(w, r, response)
}

func appWishlistAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := helpers.RegexIntsOnly.FindString(chi.URLParam(r, "id"))
	if id == "" {
		return
	}

	var hc influx.HighChartsJSON

	item := memcache.ItemAppWishlistChart(id)
	err := memcache.Client().GetSet(item.Key, item.Expiration, &hc, func() (interface{}, error) {

		builder := influxql.NewBuilder()
		builder.AddSelect("MEAN(wishlist_avg_position)", "mean_wishlist_avg_position")
		builder.AddSelect("MEAN(wishlist_count)", "mean_wishlist_count")
		builder.AddSelect("MEAN(wishlist_percent)", "mean_wishlist_percent")
		builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
		builder.AddWhere("app_id", "=", id)
		builder.AddGroupByTime("1d")
		builder.SetFillNone()

		resp, err := influx.InfluxQuery(builder)
		if err != nil {
			log.Err(err.Error(), zap.String("query", builder.String()))
			return hc, err
		}

		var hc influx.HighChartsJSON

		if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

			hc = influx.InfluxResponseToHighCharts(resp.Results[0].Series[0], true)
		}

		return hc, err
	})

	if err != nil {
		log.ErrS(err)
	}

	returnJSON(w, r, hc)
}

func appPlayersHeatmapAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := helpers.RegexIntsOnly.FindString(chi.URLParam(r, "id"))
	if id == "" {
		return
	}

	var hc = influx.HighChartsJSON{}
	var callback = func() (interface{}, error) {

		builder := influxql.NewBuilder()
		builder.AddSelect("MAX(player_count)", "max_player_count")
		builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
		builder.AddWhere("time", ">", "NOW()-28d")
		builder.AddWhere("app_id", "=", id)
		builder.AddGroupByTime("1h")
		builder.SetFillNumber(0)

		resp, err := influx.InfluxQuery(builder)
		if err != nil {
			log.Err(err.Error(), zap.String("query", builder.String()))
			return hc, err
		}

		var data = map[time.Weekday]map[int][]float64{}

		if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

			var series = resp.Results[0].Series[0]

			for k := range series.Columns {
				if k > 0 {

					for _, vv := range series.Values {

						t, err := time.Parse(time.RFC3339, vv[0].(string))
						if err != nil {
							log.ErrS(err)
							continue
						}

						val, err := vv[k].(json.Number).Float64()
						if err != nil {
							log.ErrS(err)
							continue
						}

						if data[t.Weekday()] == nil {
							data[t.Weekday()] = map[int][]float64{}
						}

						data[t.Weekday()][t.Hour()] = append(data[t.Weekday()][t.Hour()], val)
					}
				}
			}
		}
		for day, hours := range data {
			for hour, vals := range hours {
				hc["max_player_count"] = append(hc["max_player_count"], []interface{}{hour, day, helpers.Avg(vals...)})
			}
		}

		return hc, err
	}

	item := memcache.ItemAppPlayersHeatmapChart(id)
	err := memcache.Client().GetSet(item.Key, item.Expiration, &hc, callback)
	if err != nil {
		log.ErrS(err)
	}

	returnJSON(w, r, hc)
}

func appTagsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(helpers.RegexIntsOnly.FindString(chi.URLParam(r, "id")))
	if err != nil {
		return
	}

	var hc influx.HighChartsJSON
	var tagsMap = map[int]string{}
	var tagsOrder []int

	app, err := mongo.GetApp(id)
	if err != nil {
		log.ErrS(err)
		return
	}

	for _, v := range app.TagCounts {
		tagsMap[v.ID] = v.Name
		tagsOrder = append(tagsOrder, v.ID)
	}

	item := memcache.ItemAppTagsChart(id)
	err = memcache.Client().GetSet(item.Key, item.Expiration, &hc, func() (interface{}, error) {

		builder := influxql.NewBuilder()
		for _, v := range app.TagCounts {
			builder.AddSelect("MAX(tag_"+strconv.Itoa(v.ID)+")", "tag_"+strconv.Itoa(v.ID))
		}
		builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
		builder.AddWhere("app_id", "=", id)
		builder.AddGroupByTime("1d")
		builder.SetFillNone()

		resp, err := influx.InfluxQuery(builder)
		if err != nil {
			log.Err(err.Error(), zap.String("query", builder.String()))
			return hc, err
		}

		if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

			hc = influx.InfluxResponseToHighCharts(resp.Results[0].Series[0], true)
		}

		return hc, err
	})

	if err != nil {
		log.ErrS(err)
	}

	returnJSON(w, r, map[string]interface{}{
		"counts": hc,
		"names":  tagsMap,
		"order":  tagsOrder,
	})
}

func appPlayersAjaxHandler(limit bool) func(http.ResponseWriter, *http.Request) {

	var group string
	var days string
	var rolling string

	if limit {
		group = "10m"
		days = "8d"     // Gets trimmed to 7 in JS
		rolling = "144" // 1 day
	} else {
		group = "1d"
		days = "1825d"
		rolling = "7" // 1 week
	}

	return func(w http.ResponseWriter, r *http.Request) {

		id := helpers.RegexIntsOnly.FindString(chi.URLParam(r, "id"))
		if id == "" {
			return
		}

		var hc influx.HighChartsJSON
		var callback = func() (interface{}, error) {

			builder := influxql.NewBuilder()
			builder.AddSelect("MAX(player_count)", "max_player_count")
			// builder.AddSelect("max(twitch_viewers)", "max_twitch_viewers")
			if limit || session.IsLoggedIn(r) {
				builder.AddSelect("MAX(youtube_views)", "max_youtube_views")
				builder.AddSelect("MAX(youtube_comments)", "max_youtube_comments")
			}
			builder.AddSelect("MOVING_AVERAGE(max(\"player_count\"), "+rolling+")", "max_moving_average")
			builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
			builder.AddWhere("time", ">", "NOW()-"+days)
			builder.AddWhere("app_id", "=", id)
			builder.AddGroupByTime(group)
			builder.SetFillNone()

			resp, err := influx.InfluxQuery(builder)
			if err != nil {
				log.Err(err.Error(), zap.String("query", builder.String()))
				return hc, err
			}

			if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

				hc = influx.InfluxResponseToHighCharts(resp.Results[0].Series[0], !limit)
			}

			return hc, err
		}

		item := memcache.ItemAppPlayersChart(id, limit)
		err := memcache.Client().GetSet(item.Key, item.Expiration, &hc, callback)
		if err != nil {
			log.ErrS(err)
		}

		returnJSON(w, r, hc)
	}
}

// Player ranks table
func appTimeAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.ErrS(err)
		return
	}

	query := datatable.NewDataTableQuery(r, true)

	playerAppFilter := bson.D{
		{Key: "app_id", Value: id}, // needed for count too
		{Key: "app_time", Value: bson.M{"$gt": 0}},
	}

	playerApps, err := mongo.GetPlayerAppsByApp(id, query.GetOffset64(), playerAppFilter)
	if err != nil {
		log.ErrS(err)
		return
	}

	if len(playerApps) < 1 {
		return
	}

	var playerIDsMap = map[int64]int{}
	var playerIDsSlice []int64
	for _, v := range playerApps {
		playerIDsMap[v.PlayerID] = v.AppTime
		playerIDsSlice = append(playerIDsSlice, v.PlayerID)
	}

	//
	var wg sync.WaitGroup

	// Get players
	var playersAppRows []appTimeAjax
	wg.Add(1)
	go func() {

		defer wg.Done()

		players, err := mongo.GetPlayersByID(playerIDsSlice, bson.M{"_id": 1, "persona_name": 1, "avatar": 1, "country_code": 1})
		if err != nil {
			log.ErrS(err)
			return
		}

		for _, player := range players {

			if _, ok := playerIDsMap[player.ID]; !ok {
				continue
			}

			playersAppRows = append(playersAppRows, appTimeAjax{
				ID:      player.ID,
				Name:    player.GetName(),
				Avatar:  player.Avatar,
				Time:    playerIDsMap[player.ID],
				Country: player.CountryCode,
			})
		}

		sort.Slice(playersAppRows, func(i, j int) bool {
			return playersAppRows[i].Time > playersAppRows[j].Time
		})

		for k := range playersAppRows {
			playersAppRows[k].Rank = query.GetOffset() + k + 1
		}
	}()

	// Get total
	var total int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionPlayerApps, playerAppFilter, 0)
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Wait
	wg.Wait()

	response := datatable.NewDataTablesResponse(r, query, total, total, nil)
	for _, v := range playersAppRows {

		response.AddRow([]interface{}{
			strconv.FormatInt(v.ID, 10),          // 0
			v.Name,                               // 1
			helpers.GetTimeLong(v.Time, 3),       // 2
			helpers.GetPlayerFlagPath(v.Country), // 3
			helpers.OrdinalComma(v.Rank),         // 4
			helpers.GetPlayerAvatar(v.Avatar),    // 5
			helpers.GetPlayerPath(v.ID, v.Name),  // 6
			i18n.CountryCodeToName(v.Country),    // 7
		})
	}

	returnJSON(w, r, response)
}

type appTimeAjax struct {
	ID      int64
	Name    string
	Avatar  string
	Time    int
	Rank    int
	Country string
}

// Review score over time chart
func appReviewsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		return
	}

	builder := influxql.NewBuilder()
	builder.AddSelect("MEAN(reviews_score)", "mean_reviews_score")
	builder.AddSelect("MEAN(reviews_positive)", "mean_reviews_positive")
	builder.AddSelect("MEAN(reviews_negative)", "mean_reviews_negative")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW()-365d")
	builder.AddWhere("app_id", "=", id)
	builder.AddGroupByTime("1d")
	builder.SetFillNone()

	resp, err := influx.InfluxQuery(builder)
	if err != nil {
		log.Err(err.Error(), zap.String("query", builder.String()))
		return
	}

	var hc influx.HighChartsJSON

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		// Update negative numbers to be negative
		var negCol int
		for k, v := range resp.Results[0].Series[0].Columns {
			if strings.HasSuffix(v, "negative") {
				negCol = k
			}
		}
		if negCol > 0 {
			for k, v := range resp.Results[0].Series[0].Values {
				if len(v) >= negCol {
					if val, ok := v[negCol].(json.Number); ok {
						i, err := val.Float64()
						if err == nil {
							resp.Results[0].Series[0].Values[k][negCol] = json.Number(helpers.FloatToString(-i, 2))
						}
					}
				}
			}
		}

		hc = influx.InfluxResponseToHighCharts(resp.Results[0].Series[0], true)
	}

	returnJSON(w, r, hc)
}

func friendsJSONHandler(w http.ResponseWriter, r *http.Request) {

	//goland:noinspection GoPreferNilSlice
	var ret = []helpers.Tuple{}

	defer func() {
		b, _ := json.Marshal(ret)
		_, _ = w.Write(b)
	}()

	playerID := session.GetPlayerIDFromSesion(r)
	if playerID == 0 {
		return
	}

	appID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || appID == 0 {
		return
	}

	var friends []mongo.PlayerFriend
	var callback = func() (interface{}, error) {

		// Get player's friends
		friends, err := mongo.GetFriends(playerID, 0, 0, nil, bson.D{{Key: "name", Value: bson.M{"$ne": ""}}})
		if err != nil {
			return nil, err
		}

		var friendIDs []int64
		for _, v := range friends {
			friendIDs = append(friendIDs, v.FriendID)
		}

		// Make map of playerApp's
		playerApps, err := mongo.GetPlayerAppsByPlayersAndApp(friendIDs, appID)
		if err != nil {
			return nil, err
		}

		playerAppsMap := map[int64]struct{}{}
		for _, v := range playerApps {
			playerAppsMap[v.PlayerID] = struct{}{}
		}

		// Filter friends not in the map
		var friends2 []mongo.PlayerFriend
		for _, v := range friends {
			if _, ok := playerAppsMap[v.FriendID]; ok {
				friends2 = append(friends2, v)
			}
		}

		return friends2, err
	}

	item := memcache.ItemPlayerFriends(playerID, appID)
	err = memcache.Client().GetSet(item.Key, item.Expiration, &friends, callback)
	if err != nil {
		log.ErrS(err)
		return
	}

	for _, v := range friends {
		ret = append(ret, helpers.Tuple{Key: strconv.FormatInt(v.FriendID, 10), Value: v.GetName()})
	}
}
