package pages

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

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/i18n"
	"github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/sql/pics"
	"github.com/go-chi/chi"
	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
)

func appRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", appHandler)
	r.Get("/localization.html", appLocalizationHandler)
	r.Get("/reviews.html", appReviewsHandler)
	r.Get("/news.json", appNewsAjaxHandler)
	r.Get("/prices.json", appPricesAjaxHandler)
	r.Get("/players.json", appPlayersAjaxHandler)
	r.Get("/players2.json", appPlayers2AjaxHandler)
	r.Get("/items.json", appItemsAjaxHandler)
	r.Get("/reviews.json", appReviewsAjaxHandler)
	r.Get("/time.json", appTimeAjaxHandler)
	r.Get("/achievements.json", appAchievementsAjaxHandler)
	r.Get("/dlc.json", appDLCAjaxHandler)

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
		log.Warning(err)
		err = nil
	}
	if err != nil {

		if err == mongo.ErrNoDocuments {
			returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Sorry but we can not find this app."})
			return
		}

		err = helpers.IgnoreErrors(err, mongo.ErrInvalidAppID)
		log.Err(err, r)

		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the app."})
		return
	}

	// Template
	t := appTemplate{}
	t.setBackground(app, false, false)
	t.fill(w, r, app.GetName(), template.HTML(app.ShortDescription))
	t.addAssetHighCharts()
	t.addAssetJSON2HTML()
	t.metaImage = app.GetMetaImage()
	t.IncludeSocialJS = true
	t.App = app
	t.Description = template.HTML(app.ShortDescription)
	t.Canonical = app.GetPath()

	t.GroupPath = helpers.GetGroupPath(app.GroupID, app.GetName())

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

		err = queue.ProduceSteam(queue.SteamMessage{AppIDs: []int{app.ID}})
		if err == nil {
			t.addToast(Toast{Title: "Update", Message: "App has been queued for an update"})
		}
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Tags
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Tags, err = GetAppTags(app)
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Categories
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Categories, err = GetAppCategories(app)
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Genres
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Genres, err = GetAppGenres(app)
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Bundles
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Bundles, err = GetAppBundles(app)
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Get packages
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Packages, err = app.GetAppPackages()
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Get related apps
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Related, err = app.GetAppRelatedApps()
		if err != nil {
			log.Err(err, r)
			return
		}

		var tagIDs []int
		for _, v := range t.Related {
			for _, vv := range v.Tags {
				if helpers.SliceHasInt(app.Tags, vv) {
					tagIDs = append(tagIDs, vv)
				}
			}
		}

		tags, err := sql.GetTagsByID(tagIDs, []string{"id", "name"})
		if err != nil {
			log.Err(err, r)
			return
		}

		t.RelatedTags = map[int]sql.Tag{}
		for _, v := range tags {
			t.RelatedTags[v.ID] = v
		}
	}()

	// Get demos
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Demos, err = app.GetDemos()
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Get Developers
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Developers, err = GetDevelopers(app)
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Get Publishers
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Publishers, err = GetPublishers(app)
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Get played time
	wg.Add(1)
	var playedMessage string
	go func() {

		defer wg.Done()

		if session.IsLoggedIn(r) {

			playerID, err := session.GetPlayerIDFromSesion(r)
			if err == nil && playerID > 0 {

				playerApp, err := mongo.GetPlayerAppByKey(playerID, app.ID)
				if err != nil {
					err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
					log.Err(err, r)
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

	//
	t.PlayersInGame, err = t.App.GetPlayersInGame()
	if err != nil {
		log.Err(err, r)
	}

	// Functions that get called multiple times in the template
	t.Price = app.Prices.Get(session.GetProductCC(r))
	t.Common = app.Common.Formatted(app.ID, pics.CommonKeys)
	t.Extended = app.Extended.Formatted(app.ID, pics.ExtendedKeys)
	t.Config = app.Config.Formatted(app.ID, pics.ConfigKeys)
	t.UFS = app.UFS.Formatted(app.ID, pics.UFSKeys)

	//
	sort.Slice(app.Reviews.Reviews, func(i, j int) bool {
		return app.Reviews.Reviews[i].VotesGood > app.Reviews.Reviews[j].VotesGood
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
			Icon: "https://static.twitchcdn.net/assets/favicon-32-d6025c14e900565d6177.png",
			Hide: app.TwitchURL == "",
		},
		{
			Text: "View " + app.GetTypeLower() + " on Steam Prices",
			Link: "https://steamprices.com/" + app.GetSteamPricesURL() + "/" + strconv.Itoa(app.ID) + "/" + slug.Make(app.GetName()),
			Icon: "https://pbs.twimg.com/profile_images/598791990084575232/KonQ1bk8_400x400.png",
			Hide: app.GetSteamPricesURL() == "",
		},
		{
			Text: "View " + app.GetTypeLower() + " on Achievement Stats",
			Link: "https://www.achievementstats.com/index.php?action=games&gameId=" + strconv.Itoa(app.ID),
			Icon: "https://www.achievementstats.com/templates/classic/images/favicon.ico",
		},
		{
			Text: "View " + app.GetTypeLower() + " on Steam Hunters",
			Link: "https://steamhunters.com/stats/" + strconv.Itoa(app.ID) + "/achievements",
			Icon: "https://steamhunters.com/favicon-32x32.png?v=201705192248",
		},
		{
			Text: "View " + app.GetTypeLower() + " on IsThereAnyDeal",
			Link: "https://isthereanydeal.com/steam/app/" + strconv.Itoa(app.ID),
			Icon: "https://d2uym1p5obf9p8.cloudfront.net/images/favicon.png",
		},
		{
			Text: "Find similar " + app.GetTypeLower() + "s on SteamPeek",
			Link: "https://steampeek.hu/?appid=" + strconv.Itoa(app.ID),
			Icon: "https://steampeek.hu/favicon-32x32-spg.png",
		},
	}

	//
	returnTemplate(w, r, "app", t)
}

type appTemplate struct {
	GlobalTemplate
	App           mongo.App
	Banners       map[string][]string
	Bundles       []sql.Bundle
	Categories    []sql.Category
	Common        []pics.KeyValue
	Config        []pics.KeyValue
	Demos         []mongo.App
	Related       []mongo.App
	RelatedTags   map[int]sql.Tag
	Developers    []sql.Developer
	Extended      []pics.KeyValue
	Genres        []sql.Genre
	Links         []appLinkTemplate
	Packages      []mongo.Package
	Price         helpers.ProductPrice
	Publishers    []sql.Publisher
	Tags          []sql.Tag
	TagsMax       int
	UFS           []pics.KeyValue
	PlayersInGame int64
	GroupPath     string
}

type appLinkTemplate struct {
	Text string
	Link string
	Icon string
	Hide bool
}

func (t appTemplate) GetRelatedTags(relatedApp mongo.App) template.HTML {

	var ret []string
	for _, v := range relatedApp.Tags {
		if val, ok := t.RelatedTags[v]; ok {
			ret = append(ret, `<a href="`+val.GetPath()+`">`+val.GetName()+`</a>`)
		}
	}

	return template.HTML(strings.Join(ret, ", "))
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
	err = mongo.FindOne(mongo.CollectionApps, bson.D{{"_id", id}}, nil, bson.M{"localization": 1}, &app)
	if err != nil {
		err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
		log.Err(err, r)
		return
	}

	t := appLocalizationTemplate{}
	t.App = app

	returnTemplate(w, r, "app_localization", t)
}

type appLocalizationTemplate struct {
	GlobalTemplate
	App mongo.App
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
	err = mongo.FindOne(mongo.CollectionApps, bson.D{{"_id", id}}, nil, bson.M{"reviews": 1}, &app)
	if err != nil {
		err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
		log.Err(err, r)
		return
	}

	t := appReviewsTemplate{}
	t.App = app

	returnTemplate(w, r, "app_reviews", t)
}

type appReviewsTemplate struct {
	GlobalTemplate
	App mongo.App
}

func appNewsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid App ID"})
		return
	}

	query := datatable.NewDataTableQuery(r, false)

	//
	var wg sync.WaitGroup

	// Get articles
	var articles []mongo.Article
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		articles, err = mongo.GetArticlesByApp(id, query.GetOffset64())
		if err != nil {
			log.Err(err, r, id)
			return
		}
	}()

	// Get total
	var total int
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		app, err := mongo.GetApp(id)
		if err != nil {
			log.Err(err, r, id)
			return
		}

		total = len(app.NewsIDs)
	}()

	wg.Wait()

	//
	var response = datatable.NewDataTablesResponse(r, query, int64(total), int64(total), nil)
	for _, article := range articles {

		var id = strconv.FormatInt(article.ID, 10)
		var path = helpers.GetAppPath(article.AppID, article.AppName)

		response.AddRow([]interface{}{
			id,                                    // 0
			article.Title,                         // 1
			article.Author,                        // 2
			article.Date.Unix(),                   // 3
			article.Date.Format(helpers.DateYear), // 4
			article.GetBody(),                     // 5
			article.AppID,                         // 6
			article.AppName,                       // 7
			article.GetIcon(),                     // 8
			path + "#news," + id,                  // 9
			helpers.DefaultAppIcon,                // 10
		})
	}

	returnJSON(w, r, response)
}

func appPricesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	productPricesAjaxHandler(w, r, helpers.ProductTypeApp)
}

func appAchievementsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return
	}

	var query = datatable.NewDataTableQuery(r, false)
	var filter = bson.D{{"app_id", id}}

	//
	var wg sync.WaitGroup

	// Get achievements
	var achievements []mongo.AppAchievement
	var achievedMap = map[string]bool{}
	wg.Add(1)
	go func() {

		defer wg.Done()

		var sortOrder = query.GetOrderMongo(map[string]string{
			"1": "completed",
		})

		var err error
		achievements, err = mongo.GetAppAchievements(query.GetOffset64(), 100, filter, sortOrder)
		if err != nil {
			log.Err(err, r)
			return
		}

		playerID, err := session.GetPlayerIDFromSesion(r)
		if err == nil && playerID > 0 {

			var a = bson.A{}
			for _, v := range achievements {
				a = append(a, v.Key)
			}

			playerAchievements, err := mongo.GetPlayerAchievementsForApp(playerID, id, a, 0)
			if err != nil {
				log.Err(err, r)
				return
			}

			for _, v := range playerAchievements {
				achievedMap[v.AchievementID] = true
			}
		}
	}()

	// Get total
	var total int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionAppAchievements, filter, 60*60*24*28)
		if err != nil {
			log.Err(err, r)
			return
		}
	}()

	// Wait
	wg.Wait()

	response := datatable.NewDataTablesResponse(r, query, total, total, nil)
	for _, achievement := range achievements {

		_, achieved := achievedMap[achievement.Key]

		completed := helpers.FloatToString(achievement.Completed, 1)

		response.AddRow([]interface{}{
			achievement.Name,        // 0
			achievement.Description, // 1
			achievement.GetIcon(),   // 2
			completed,               // 3
			achievement.Active,      // 4
			achievement.Hidden,      // 5
			achievement.Deleted,     // 6
			achieved,                // 7
		})
	}

	returnJSON(w, r, response)
}

func appDLCAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return
	}

	var query = datatable.NewDataTableQuery(r, false)
	var search = query.GetSearchString("search")

	var filter = bson.D{{"app_id", id}}
	var filter2 = filter

	if len(search) > 1 {

		quoted := regexp.QuoteMeta(search)

		filter2 = append(filter, bson.E{Key: "name", Value: bson.M{"$regex": quoted, "$options": "i"}})
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
		dlcs, err = mongo.GetDLCForApp(query.GetOffset64(), 100, filter2, sortOrder)
		if err != nil {
			log.Err(err, r)
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
			log.Err(err, r)
			return
		}

		filtered, err = mongo.CountDocuments(mongo.CollectionAppDLC, filter2, 60*60*24)
		if err != nil {
			log.Err(err, r)
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
		log.Err(err, r)
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
			log.Err(err, r)
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
			log.Err(err, r)
		}

		filtered, err = mongo.CountDocuments(mongo.CollectionAppItems, filter2, 0)
		if err != nil {
			log.Err(err, r)
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

// Player counts chart
func appPlayersAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		return
	}

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.AddSelect("max(twitch_viewers)", "max_twitch_viewers")
	builder.AddSelect("max(youtube_views)", "max_youtube_views")
	builder.AddSelect("max(youtube_comments)", "max_youtube_comments")
	builder.AddSelect("MOVING_AVERAGE(max(\"player_count\"), 20)", "max_moving_average")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW()-14d")
	builder.AddWhere("app_id", "=", id)
	builder.AddGroupByTime("10m")
	builder.SetFillNone()

	resp, err := influx.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var hc influx.HighChartsJSON

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		hc = influx.InfluxResponseToHighCharts(resp.Results[0].Series[0])
	}

	returnJSON(w, r, hc)
}

// Player counts chart - 1 year
func appPlayers2AjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		return
	}

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.AddSelect("max(twitch_viewers)", "max_twitch_viewers")
	builder.AddSelect("max(youtube_views)", "max_youtube_views")
	builder.AddSelect("max(youtube_comments)", "max_youtube_comments")
	builder.AddSelect("MOVING_AVERAGE(max(\"player_count\"), 20)", "max_moving_average")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW()-1825d")
	builder.AddWhere("app_id", "=", id)
	builder.AddGroupByTime("1d")
	builder.SetFillNone()

	resp, err := influx.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var hc influx.HighChartsJSON

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		hc = influx.InfluxResponseToHighCharts(resp.Results[0].Series[0])
	}

	returnJSON(w, r, hc)
}

// Player ranks table
func appTimeAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Err(err, r)
		return
	}

	query := datatable.NewDataTableQuery(r, true)

	playerAppFilter := bson.D{
		{Key: "app_id", Value: id},
		{Key: "app_time", Value: bson.M{"$gt": 0}},
	}

	playerApps, err := mongo.GetPlayerAppsByApp(query.GetOffset64(), playerAppFilter)
	if err != nil {
		log.Err(err, r)
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
			log.Err(err, r)
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
		log.Err(err, r)
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
	builder.AddSelect("mean(reviews_score)", "mean_reviews_score")
	builder.AddSelect("mean(reviews_positive)", "mean_reviews_positive")
	builder.AddSelect("mean(reviews_negative)", "mean_reviews_negative")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW()-365d")
	builder.AddWhere("app_id", "=", id)
	builder.AddGroupByTime("1d")
	builder.SetFillNone()

	resp, err := influx.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
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

		hc = influx.InfluxResponseToHighCharts(resp.Results[0].Series[0])
	}

	returnJSON(w, r, hc)
}
