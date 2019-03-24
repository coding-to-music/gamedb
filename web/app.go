package web

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/mongo"
	"github.com/gamedb/website/queue"
	"github.com/gamedb/website/session"
	"github.com/go-chi/chi"
)

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
	app, err := db.GetApp(idx, []string{})
	if err != nil {

		if err == db.ErrRecordNotFound {
			returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Sorry but we can not find this app."})
			return
		}

		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the app.", Error: err})
		return
	}

	// Template
	t := appTemplate{}
	t.fill(w, r, app.GetName(), "")
	t.MetaImage = app.GetMetaImage()
	t.addAssetCarousel()
	t.addAssetHighCharts()
	t.App = app
	t.Description = template.HTML(app.ShortDescription)

	// Update news, reviews etc
	func() {

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

	//
	var wg sync.WaitGroup

	// Get achievements
	wg.Add(1)
	go func(app db.App) {

		defer wg.Done()

		var achievements []db.AppAchievement

		err := helpers.Unmarshal([]byte(app.Achievements), &achievements)
		log.Err(err, r)

	}(app)

	// Tags
	wg.Add(1)
	go func(app db.App) {

		defer wg.Done()

		var err error
		t.Tags, err = app.GetTags()
		log.Err(err, r)

	}(app)

	// Genres
	wg.Add(1)
	go func(app db.App) {

		defer wg.Done()

		var err error
		t.Genres, err = app.GetGenres()
		log.Err(err, r)

	}(app)

	// Bundles
	wg.Add(1)
	go func() {

		defer wg.Done()

		gorm, err := db.GetMySQLClient()
		if err != nil {
			log.Err(err, r)
			return
		}

		gorm = gorm.Where("JSON_CONTAINS(app_ids, '[" + strconv.Itoa(app.ID) + "]')")
		gorm = gorm.Find(&t.Bundles)
		if gorm.Error != nil {
			log.Err(gorm.Error, r)
			return
		}
	}()

	// Get packages
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Packages, err = db.GetPackagesAppIsIn(app.ID)
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

			gorm, err := db.GetMySQLClient()
			if err != nil {
				log.Err(err, r)
				return
			}

			var demos []db.App
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
		t.DLC, err = db.GetDLC(app, []string{"id", "name"})
		log.Err(err, r)

	}()

	// Wait
	wg.Wait()

	// Get price
	t.Price = db.GetPriceFormatted(app, session.GetCountryCode(r))

	// Functions that get called multiple times in the template
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

	// Make banners
	banners := make(map[string][]string)
	var primary []string

	if app.ID == 753 {
		primary = append(primary, "This app record is for the Steam client")
	}

	if len(t.Achievements) > 5000 {
		primary = append(primary, "Apps are limited to 5000 achievements, but this app created more before the limit put in place")
	}

	if len(primary) > 0 {
		banners["primary"] = primary
	}

	t.Banners = banners

	//
	err = returnTemplate(w, r, "app", t)
	log.Err(err, r)
}

type appTemplate struct {
	GlobalTemplate
	Achievements []db.AppAchievement
	App          db.App
	Banners      map[string][]string
	Bundles      []db.Bundle
	Demos        []db.App
	Developers   []db.Developer
	DLC          []db.App
	Genres       []db.Genre
	Movies       []db.AppVideo
	NewsIDs      []int64
	Packages     []db.Package
	Price        db.ProductPriceFormattedStruct
	Prices       db.ProductPrices
	Publishers   []db.Publisher
	Reviews      db.AppReviewSummary
	Screenshots  []db.AppImage
	SteamSpy     db.AppSteamSpy
	Stats        []db.AppStat
	Tags         []db.Tag
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
	log.Err(err, r, idx)

	//
	var wg sync.WaitGroup

	// Get events
	var articles []mongo.Article

	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		articles, err := mongo.GetArticles(idx, query.getOffset64())
		log.Err(err, r, idx)

		// todo, add http to links here instead of JS
		// var regex = regexp.MustCompile(`href="(?!http)(.*)"`)
		// var conv bbConvert.HTMLConverter
		// conv.ImplementDefaults()
		// Fix broken links
		// v.Contents = regex.ReplaceAllString(v.Contents, `$1http://$2`)
		// Convert BBCdoe to HTML
		// v.Contents = conv.Convert(v.Contents)

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
		app, err := db.GetApp(idx, []string{})
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
	response.RecordsTotal = strconv.Itoa(total)
	response.RecordsFiltered = strconv.Itoa(total)
	response.Draw = query.Draw

	for _, v := range articles {
		response.AddRow(v.OutputForJSON())
	}

	response.output(w, r)
}

func appPricesAjaxHandler(w http.ResponseWriter, r *http.Request) {
	productPricesAjaxHandler(w, r, helpers.ProductTypeApp)
}

func appPlayersAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Err("invalid id", r)
		return
	}

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.AddSelect("max(twitch_viewers)", "max_twitch_viewers")
	builder.SetFrom("GameDB", "alltime", "apps")
	builder.AddWhere("time", ">", "NOW()-7d")
	builder.AddWhere("app_id", "=", id)
	builder.AddGroupByTime("10m")
	builder.SetFillNone()

	resp, err := db.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var hc db.HighChartsJson

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		hc = db.InfluxResponseToHighCharts(resp.Results[0].Series[0])
	}

	b, err := json.Marshal(hc)
	if err != nil {
		log.Err(err, r)
		return
	}

	err = returnJSON(w, r, b)
	if err != nil {
		log.Err(err, r)
		return
	}
}

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
	builder.SetFrom("GameDB", "alltime", "apps")
	builder.AddWhere("time", ">", "NOW()-365d")
	builder.AddWhere("app_id", "=", id)
	builder.AddGroupByTime("1d")
	builder.SetFillNone()

	resp, err := db.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var hc db.HighChartsJson

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		hc = db.InfluxResponseToHighCharts(resp.Results[0].Series[0])
	}

	b, err := json.Marshal(hc)
	if err != nil {
		log.Err(err, r)
		return
	}

	err = returnJSON(w, r, b)
	if err != nil {
		log.Err(err, r)
		return
	}
}
