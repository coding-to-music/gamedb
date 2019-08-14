package pages

import (
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/sql/pics"
	"github.com/go-chi/chi"
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
	app, err := sql.GetApp(idx, nil)
	if err != nil {

		if err == sql.ErrRecordNotFound {
			returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Sorry but we can not find this app."})
			return
		}

		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the app.", Error: err})
		return
	}

	// Template
	t := appTemplate{}
	t.setBackground(app, false, false)
	t.fill(w, r, app.GetName(), "")
	t.metaImage = app.GetMetaImage()
	t.addAssetCarousel()
	t.addAssetHighCharts()
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

		if app.UpdatedAt.Unix() > time.Now().Add(time.Hour * -24).Unix() {
			return
		}

		err = queue.ProduceApp(app.ID)
		if err != nil {
			log.Err(err, r)
		} else {
			t.addToast(Toast{Title: "Update", Message: "App has been queued for an update"})
		}
	}()

	// Tags
	wg.Add(1)
	go func(app sql.App) {

		defer wg.Done()

		var err error
		t.Tags, err = app.GetTags()
		log.Err(err, r)
	}(app)

	// // Categories
	// wg.Add(1)
	// go func(app sql.App) {
	//
	// 	defer wg.Done()
	//
	// 	var err error
	// 	t.Tags, err = app.GetCategoryIDs()
	// 	log.Err(err, r)
	// }(app)

	// Genres
	wg.Add(1)
	go func(app sql.App) {

		defer wg.Done()

		var err error
		t.Genres, err = app.GetGenres()
		log.Err(err, r)
	}(app)

	// Bundles
	wg.Add(1)
	go func() {

		defer wg.Done()

		gorm, err := sql.GetMySQLClient()
		if err != nil {
			log.Err(err, r)
			return
		}

		gorm = gorm.Where("JSON_CONTAINS(app_ids, '[" + strconv.Itoa(app.ID) + "]')")
		gorm = gorm.Find(&t.Bundles)

		log.Err(gorm.Error, r)
	}()

	// Get packages
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Packages, err = sql.GetPackagesAppIsIn(app.ID)
		log.Err(err, r)
	}()

	// Get demos
	wg.Add(1)
	go func() {

		defer wg.Done()

		demoIDs, err := app.GetDemoIDs()
		if err != nil {
			log.Err(err, r)
			return
		}

		if len(demoIDs) > 0 {

			gorm, err := sql.GetMySQLClient()
			if err != nil {
				log.Err(err, r)
				return
			}

			var demos []sql.App
			gorm = gorm.Where("id IN (?)", demoIDs)
			gorm = gorm.Find(&demos)
			if gorm.Error != nil {
				log.Err(gorm.Error, r)
				return
			}

			t.Demos = demos
		}
	}()

	// Get DLC
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.DLCs, err = sql.GetDLC(app, []string{"id", "name"})
		log.Err(err, r)
	}()

	// Wait
	wg.Wait()

	// Functions that get called multiple times in the template
	t.Price = app.GetPrice(helpers.GetProductCC(r))

	t.Achievements, err = t.App.GetAchievements()
	log.Err(err, r)

	t.NewsIDs, err = t.App.GetNewsIDs()
	log.Err(err, r)

	t.Stats, err = t.App.GetStats()
	log.Err(err, r)

	t.Prices, err = t.App.GetPrices()
	log.Err(err, r)

	t.Screenshots, err = t.App.GetScreenshots()
	log.Err(err, r)

	t.Movies, err = t.App.GetMovies()
	log.Err(err, r)

	t.Reviews, err = t.App.GetReviews()
	log.Err(err, r)

	t.Developers, err = t.App.GetDevelopers()
	log.Err(err, r)

	t.Publishers, err = t.App.GetPublishers()
	log.Err(err, r)

	t.SteamSpy, err = t.App.GetSteamSpy()
	log.Err(err, r)

	t.Common, err = t.App.GetCommon().Formatted(app.ID, pics.CommonKeys)
	log.Err(err, r)

	t.Extended, err = t.App.GetExtended().Formatted(app.ID, pics.ExtendedKeys)
	log.Err(err, r)

	t.Config, err = t.App.GetConfig().Formatted(app.ID, pics.ConfigKeys)
	log.Err(err, r)

	t.UFS, err = t.App.GetUFS().Formatted(app.ID, pics.UFSKeys)
	log.Err(err, r)

	// Make banners
	var banners = map[string][]string{
		"primary": []string{},
		"warning": []string{},
	}

	if app.ID == 753 {
		banners["primary"] = append(banners["primary"], "This app record is for the Steam client")
	}

	if app.GetCommon().GetValue("app_retired_publisher_request") == "1" {
		banners["warning"] = append(banners["warning"], "At the request of the publisher, "+app.GetName()+" is no longer available for sale on Steam.")
	}

	t.Banners = banners

	//
	err = returnTemplate(w, r, "app", t)
	log.Err(err, r)
}

type appTemplate struct {
	GlobalTemplate
	Achievements []sql.AppAchievement
	App          sql.App
	Banners      map[string][]string
	Bundles      []sql.Bundle
	Categories   []sql.Category
	Common       []pics.KeyValue
	Config       []pics.KeyValue
	Demos        []sql.App
	Developers   []sql.Developer
	DLCs         []sql.App
	Extended     []pics.KeyValue
	Genres       []sql.Genre
	Movies       []sql.AppVideo
	NewsIDs      []int64
	Packages     []sql.Package
	Price        sql.ProductPrice
	Prices       sql.ProductPrices
	Publishers   []sql.Publisher
	Reviews      sql.AppReviewSummary
	Screenshots  []sql.AppImage
	SteamSpy     sql.AppSteamSpy
	Stats        []sql.AppStat
	Tags         []sql.Tag
	UFS          []pics.KeyValue
}

func (t appTemplate) GetReleaseDate() string {
	nice := t.App.GetReleaseDateNice()
	state := t.App.GetReleaseState()

	if nice != "" {
		state = " (" + state + ")"
	}

	return nice + state
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

	query := DataTablesQuery{}
	err = query.fillFromURL(r.URL.Query())
	if err != nil {
		log.Err(err, r, idx)
	}

	query.limit(r)

	//
	var wg sync.WaitGroup

	// Get events
	var articles []mongo.Article

	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		articles, err = mongo.GetArticlesByApp(idx, query.getOffset64())
		if err != nil {
			log.Err(err, r, idx)
			return
		}

		for k, v := range articles {
			articles[k].Contents = helpers.BBCodeCompiler.Compile(v.Contents)
		}
	}(r)

	// Get total
	var total int
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		app, err := sql.GetApp(idx, nil)
		if err != nil {
			log.Err(err, r, idx)
			return
		}

		newsIDs, err := app.GetNewsIDs()
		if err != nil {
			log.Err(err, r, idx)
			return
		}

		total = len(newsIDs)

	}(r)

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = int64(total)
	response.RecordsFiltered = int64(total)
	response.Draw = query.Draw
	response.limit(r)

	for _, v := range articles {
		response.AddRow(v.OutputForJSON())
	}

	response.output(w, r)
}

func appPricesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	productPricesAjaxHandler(w, r, helpers.ProductTypeApp)
}

func appItemsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	// id := chi.URLParam(r, "id")
	// if id == "" {
	// 	log.Err("invalid id", r)
	// 	return
	// }
	//
	// idx, err := strconv.Atoi(id)
	// if err != nil {
	// 	log.Err(err, r)
	// 	return
	// }
	//
	// query := DataTablesQuery{}
	// err = query.fillFromURL(r.URL.Query())
	// log.Err(err, r)
	//
	// query.limit(r)
	//
	// playerAppFilter := mongo.M{"app_id": idx, "app_time": mongo.M{"$gt": 0}}
	//
	// playerApps, err := mongo.GetPlayerAppsByApp(query.getOffset64(), playerAppFilter)
	// if err != nil {
	// 	log.Err(err, r)
	// 	return
	// }
	//
	// if len(playerApps) < 1 {
	// 	return
	// }
	//
	// var playerIDsMap = map[int64]int{}
	// var playerIDsSlice []int64
	// for _, v := range playerApps {
	// 	playerIDsMap[v.PlayerID] = v.AppTime
	// 	playerIDsSlice = append(playerIDsSlice, v.PlayerID)
	// }
	//
	// //
	// var wg sync.WaitGroup
	//
	// // Get players
	// var playersAppRows []appTimeAjax
	// wg.Add(1)
	// go func() {
	//
	// 	defer wg.Done()
	//
	// 	players, err := mongo.GetPlayersByID(playerIDsSlice, mongo.M{"_id": 1, "persona_name": 1, "avatar": 1, "country_code": 1})
	// 	if err != nil {
	// 		log.Err(err)
	// 		return
	// 	}
	//
	// 	for _, player := range players {
	//
	// 		if _, ok := playerIDsMap[player.ID]; !ok {
	// 			continue
	// 		}
	//
	// 		playersAppRows = append(playersAppRows, appTimeAjax{
	// 			ID:      player.ID,
	// 			Name:    player.PersonaName,
	// 			Avatar:  player.Avatar,
	// 			Time:    playerIDsMap[player.ID],
	// 			Country: player.CountryCode,
	// 		})
	// 	}
	//
	// 	sort.Slice(playersAppRows, func(i, j int) bool {
	// 		return playersAppRows[i].Time > playersAppRows[j].Time
	// 	})
	//
	// 	for k := range playersAppRows {
	// 		playersAppRows[k].Rank = query.getOffset() + k + 1
	// 	}
	// }()
	//
	// // Get total
	// var total int64
	// wg.Add(1)
	// go func() {
	//
	// 	defer wg.Done()
	//
	// 	var err error
	// 	total, err = mongo.CountDocuments(mongo.CollectionPlayerApps, playerAppFilter, 0)
	// 	log.Err(err, r)
	// }()
	//
	// // Wait
	// wg.Wait()
	//
	// response := DataTablesAjaxResponse{}
	// response.RecordsTotal = total
	// response.RecordsFiltered = total
	// response.Draw = query.Draw
	// response.limit(r)
	//
	// for _, v := range playersAppRows {
	//
	// 	response.AddRow([]interface{}{
	// 		strconv.FormatInt(v.ID, 10),          // 0
	// 		v.Name,                               // 1
	// 		helpers.GetTimeLong(v.Time, 3),       // 2
	// 		helpers.GetPlayerFlagPath(v.Country), // 3
	// 		helpers.OrdinalComma(v.Rank),         // 4
	// 		helpers.GetPlayerAvatar(v.Avatar),    // 5
	// 		helpers.GetPlayerPath(v.ID, v.Name),  // 6
	// 		helpers.CountryCodeToName(v.Country), // 7
	// 	})
	// }
	//
	// response.output(w, r)
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
	builder.SetFrom(helpers.InfluxGameDB, helpers.InfluxRetentionPolicyAllTime.String(), helpers.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW()-7d")
	builder.AddWhere("app_id", "=", id)
	builder.AddGroupByTime("10m")
	builder.SetFillNone()

	resp, err := helpers.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var hc helpers.HighChartsJson

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		hc = helpers.InfluxResponseToHighCharts(resp.Results[0].Series[0])
	}

	err = returnJSON(w, r, hc)
	log.Err(err, r)
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

	query := DataTablesQuery{}
	err = query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	query.limit(r)

	playerAppFilter := mongo.M{"app_id": idx, "app_time": mongo.M{"$gt": 0}}

	playerApps, err := mongo.GetPlayerAppsByApp(query.getOffset64(), playerAppFilter)
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

		players, err := mongo.GetPlayersByID(playerIDsSlice, mongo.M{"_id": 1, "persona_name": 1, "avatar": 1, "country_code": 1})
		if err != nil {
			log.Err(err)
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
			playersAppRows[k].Rank = query.getOffset() + k + 1
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

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = total
	response.RecordsFiltered = total
	response.Draw = query.Draw
	response.limit(r)

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

	response.output(w, r)
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
	builder.SetFrom(helpers.InfluxGameDB, helpers.InfluxRetentionPolicyAllTime.String(), helpers.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW()-365d")
	builder.AddWhere("app_id", "=", id)
	builder.AddGroupByTime("1d")
	builder.SetFillNone()

	resp, err := helpers.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var hc helpers.HighChartsJson

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		hc = helpers.InfluxResponseToHighCharts(resp.Results[0].Series[0])
	}

	err = returnJSON(w, r, hc)
	log.Err(err, r)
}
