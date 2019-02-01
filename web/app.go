package web

import (
	"html/template"
	"net/http"
	"strconv"
	"sync"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
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

	if !db.IsValidAppID(idx) {
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

	// Redirect to correct slug
	if r.URL.Path != app.GetPath() {
		http.Redirect(w, r, app.GetPath(), 302)
		return
	}

	// Template
	t := appTemplate{}
	t.Fill(w, r, app.GetName(), "")
	t.MetaImage = app.GetFirstScreenshot()
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

		err = queue.QueueApp([]int{app.ID})
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

	// Make banners
	banners := make(map[string][]string)
	var primary []string

	if app.ID == 753 {
		primary = append(primary, "This app record is for the Steam client.")
	}

	if len(primary) > 0 {
		banners["primary"] = primary
	}

	t.Banners = banners

	err = returnTemplate(w, r, "app", t)
	log.Err(err, r)
}

type appTemplate struct {
	GlobalTemplate
	App      db.App
	Price    db.ProductPriceFormattedStruct
	Packages []db.Package
	Bundles  []db.Bundle
	DLC      []db.App
	Tags     []db.Tag
	Genres   []db.Genre
	Banners  map[string][]string
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
	err = query.FillFromURL(r.URL.Query())
	log.Err(err, r)

	//
	var wg sync.WaitGroup

	// Get events
	var articles []db.News

	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		client, ctx, err := db.GetDSClient()
		if err != nil {

			log.Err(err, r)
			return
		}

		q := datastore.NewQuery(db.KindNews).Filter("app_id =", idx).Limit(100)
		q, err = query.SetOrderOffsetDS(q, map[string]string{})
		q = q.Order("-date")
		if err != nil {

			log.Err(err, r)
			return
		}

		_, err = client.GetAll(ctx, q, &articles)
		log.Err(err, r)

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
			log.Err(err, r)
			return
		}

		newsIDs, err := app.GetNewsIDs()
		if err != nil {
			log.Err(err, r)
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
		response.AddRow(v.OutputForJSON(r))
	}

	response.output(w, r)
}

func appPricesAjaxHandler(w http.ResponseWriter, r *http.Request) {
	productPricesAjaxHandler(w, r, db.ProductTypeApp)
}
