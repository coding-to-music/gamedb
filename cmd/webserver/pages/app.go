package pages

import (
	"encoding/json"
	"html/template"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/cmd/webserver/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/sql/pics"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

func appRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", appHandler)
	r.Get("/news.json", appNewsAjaxHandler)
	r.Get("/prices.json", appPricesAjaxHandler)
	r.Get("/players.json", appPlayersAjaxHandler)
	r.Get("/items.json", appItemsAjaxHandler)
	r.Get("/reviews.json", appReviewsAjaxHandler)
	r.Get("/time.json", appTimeAjaxHandler)
	r.Get("/{slug}", appHandler)
	return r
}

func appHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid App ID."})
		return
	}

	idx, err := strconv.Atoi(id)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid App ID: " + id})
		return
	}

	if !helpers.IsValidAppID(idx) {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid App ID: " + id})
		return
	}

	// Get app
	app, err := mongo.GetApp(idx)
	if err != nil && strings.HasPrefix(err.Error(), "memcache: unexpected response line from \"set\":") {
		log.Warning(err)
		err = nil
	}
	if err != nil {

		if err == mongo.ErrNoDocuments {
			returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Sorry but we can not find this app."})
			return
		}

		log.Err(err, r)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the app."})
		return
	}

	// Template
	t := appTemplate{}
	t.setBackground(app, false, false)
	t.fill(w, r, app.GetName(), template.HTML(app.ShortDescription))
	t.metaImage = app.GetMetaImage()
	t.addAssetCarousel()
	t.addAssetHighCharts()
	t.IncludeSocialJS = true
	t.App = app
	t.Description = template.HTML(app.ShortDescription)
	t.Canonical = app.GetPath()

	//
	var wg sync.WaitGroup

	// Update news, reviews etc
	wg.Add(1)
	go func() {

		defer wg.Done()

		if helpers.IsBot(r.UserAgent()) {
			return
		}

		if app.UpdatedAt.After(time.Now().Add(time.Hour * -24)) {
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

	// Get DLC
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.DLCs, err = app.GetDLCs()
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

	// Wait
	wg.Wait()

	// Functions that get called multiple times in the template
	t.Price = app.Prices.Get(helpers.GetProductCC(r))
	t.Common = app.Common.Formatted(app.ID, pics.CommonKeys)
	t.Extended = app.Extended.Formatted(app.ID, pics.ExtendedKeys)
	t.Config = app.Config.Formatted(app.ID, pics.ConfigKeys)
	t.UFS = app.UFS.Formatted(app.ID, pics.UFSKeys)

	//
	sort.Slice(app.Reviews.Reviews, func(i, j int) bool {
		return app.Reviews.Reviews[i].VotesGood > app.Reviews.Reviews[j].VotesGood
	})

	// Make banners
	var banners = map[string][]string{
		"primary": {},
		"warning": {},
	}

	if app.ID == 753 {
		banners["primary"] = append(banners["primary"], "This app record is for the Steam client")
	}

	if app.ReadPICS(app.Common).GetValue("app_retired_publisher_request") == "1" {
		banners["warning"] = append(banners["warning"], "At the request of the publisher, "+app.GetName()+" is no longer available for sale on Steam.")
	}

	t.Banners = banners

	//
	returnTemplate(w, r, "app", t)
}

type appTemplate struct {
	GlobalTemplate
	App         mongo.App
	Banners     map[string][]string
	Bundles     []sql.Bundle
	Categories  []sql.Category
	Common      []pics.KeyValue
	Config      []pics.KeyValue
	Demos       []mongo.App
	Related     []mongo.App
	RelatedTags map[int]sql.Tag
	Developers  []sql.Developer
	DLCs        []mongo.App
	Extended    []pics.KeyValue
	Genres      []sql.Genre
	Packages    []mongo.Package
	Price       helpers.ProductPrice
	Publishers  []sql.Publisher
	Tags        []sql.Tag
	UFS         []pics.KeyValue
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

func appNewsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid App ID."})
		return
	}

	idx, err := strconv.Atoi(id)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid App ID: " + id})
		return
	}

	query := datatable.NewDataTableQuery(r, true)

	//
	var wg sync.WaitGroup

	// Get articles
	var articles []mongo.Article
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		articles, err = mongo.GetArticlesByApp(idx, query.GetOffset64())
		if err != nil {
			log.Err(err, r, idx)
			return
		}

		for k, v := range articles {
			articles[k].Contents = helpers.BBCodeCompiler.Compile(v.Contents)
		}
	}()

	// Get total
	var total int
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		app, err := mongo.GetApp(idx)
		if err != nil {
			log.Err(err, r, idx)
			return
		}

		total = len(app.NewsIDs)
	}()

	wg.Wait()

	//
	var response = datatable.NewDataTablesResponse(r, query, int64(total), int64(total))
	for _, article := range articles {
		response.AddRow(article.OutputForJSON())
	}

	returnJSON(w, r, response)
}

func appPricesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	productPricesAjaxHandler(w, r, helpers.ProductTypeApp)
}

func appItemsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Err("invalid id", r)
		return
	}

	idx, err := strconv.Atoi(id)
	if err != nil {
		log.Err(err, r)
		return
	}

	query := datatable.NewDataTableQuery(r, true)

	// Make filter
	var search = query.GetSearchString("search")

	filter := bson.D{
		{Key: "app_id", Value: idx},
	}

	if len(search) > 1 {

		quoted := regexp.QuoteMeta(search)

		filter = append(filter, bson.E{Key: "$or", Value: bson.A{
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
		items, err = mongo.GetAppItems(query.GetOffset64(), 100, filter, nil)
		if err != nil {
			log.Err(err, r)
			return
		}

	}()

	// Get total
	var total int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionAppItems, bson.D{{Key: "app_id", Value: idx}}, 0)
		log.Err(err, r)
	}()

	// Get filtered count
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionAppItems, filter, 0)
		log.Err(err, r)
	}()

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, total, filtered)
	for _, item := range items {

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
			item.Image(36, true),    // 25
			item.Image(256, false),  // 26
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
		log.Err("invalid id", r)
		return
	}

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.AddSelect("max(twitch_viewers)", "max_twitch_viewers")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW()-7d")
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

// Player ranks table
func appTimeAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Err("invalid id", r)
		return
	}

	idx, err := strconv.Atoi(id)
	if err != nil {
		log.Err(err, r)
		return
	}

	query := datatable.NewDataTableQuery(r, true)

	playerAppFilter := bson.D{
		{Key: "app_id", Value: idx},
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
				Name:    player.PersonaName,
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

	response := datatable.NewDataTablesResponse(r, query, total, total)
	for _, v := range playersAppRows {

		response.AddRow([]interface{}{
			strconv.FormatInt(v.ID, 10),          // 0
			v.Name,                               // 1
			helpers.GetTimeLong(v.Time, 3),       // 2
			helpers.GetPlayerFlagPath(v.Country), // 3
			helpers.OrdinalComma(v.Rank),         // 4
			helpers.GetPlayerAvatar(v.Avatar),    // 5
			helpers.GetPlayerPath(v.ID, v.Name),  // 6
			helpers.CountryCodeToName(v.Country), // 7
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
		log.Err("invalid id", r)
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
