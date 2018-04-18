package web

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/CalebQ42/bbConvert"
	"github.com/go-chi/chi"
	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/steam"
)

func AppHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	idx, err := strconv.Atoi(id)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, "Invalid App ID: "+id)
		return
	}

	// Get app
	app, err := mysql.GetApp(idx)
	if err != nil {

		if err.Error() == "no id" {
			returnErrorTemplate(w, r, 404, "We can't find this app in our database, there may not be one with this ID.")
			return
		}

		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Redirect to correct slug
	if r.URL.Path != app.GetPath() {
		http.Redirect(w, r, app.GetPath(), 302)
		return
	}

	// Update news, reviews etc
	app.UpdateFromRequest(r.UserAgent())

	//
	var wg sync.WaitGroup

	var achievements []appAchievementTemplate
	wg.Add(1)
	go func() {

		// Get achievements
		achievementsResp, err := steam.GetGlobalAchievementPercentagesForApp(app.ID)
		if err != nil {
			logger.Error(err)
		}

		achievementsMap := make(map[string]float64)
		for _, v := range achievementsResp {
			achievementsMap[v.Name] = v.Percent
		}

		// Get schema
		schema, err := steam.GetSchemaForGame(app.ID)
		if err != nil {
			logger.Error(err)
		}

		// Make template struct
		for _, v := range schema.AvailableGameStats.Achievements {
			achievements = append(achievements, appAchievementTemplate{
				v.Icon,
				v.DisplayName,
				v.Description,
				achievementsMap[v.Name],
			})
		}

		wg.Done()
	}()

	var tags []mysql.Tag
	wg.Add(1)
	go func() {

		// Get tags
		tagIDs, err := app.GetTags()
		if err != nil {
			logger.Error(err)
		}

		tags, err = mysql.GetTagsByID(tagIDs)
		if err != nil {
			logger.Error(err)
		}

		wg.Done()
	}()

	var pricesString string
	var pricesCount int
	wg.Add(1)
	go func() {

		// Get prices
		pricesResp, err := datastore.GetAppPrices(app.ID, 0)
		if err != nil {
			logger.Error(err)
		}

		pricesCount = len(pricesResp)

		var prices [][]float64

		for _, v := range pricesResp {

			prices = append(prices, []float64{float64(v.CreatedAt.Unix()), float64(v.PriceFinal) / 100})
		}

		// Add current price
		prices = append(prices, []float64{float64(time.Now().Unix()), float64(app.PriceFinal) / 100})

		// Make into a JSON string
		pricesBytes, err := json.Marshal(prices)
		if err != nil {
			logger.Error(err)
		}

		pricesString = string(pricesBytes)

		wg.Done()
	}()

	var news []appArticleTemplate
	wg.Add(1)
	go func() {

		// Get news
		newsResp, err := datastore.GetArticles(idx, 1000)
		if err != nil {
			logger.Error(err)
		}

		var conv bbConvert.HTMLConverter
		conv.ImplementDefaults()

		for _, v := range newsResp {
			news = append(news, appArticleTemplate{
				Title:    v.Title,
				Contents: template.HTML(conv.Convert(v.Contents)),
				Author:   v.Author,
			})
		}

		wg.Done()
	}()

	var packages []mysql.Package
	wg.Add(1)
	go func() {

		// Get packages
		packages, err = mysql.GetPackagesAppIsIn(app.ID)
		if err != nil {
			logger.Error(err)
		}

		wg.Done()
	}()

	var dlc []mysql.App
	wg.Add(1)
	go func() {

		// Get DLC
		dlc, err = mysql.GetDLC(app, []string{"id", "name"})
		if err != nil {
			logger.Error(err)
		}

		wg.Done()
	}()

	// Make banners
	banners := make(map[string][]string)
	var primary []string

	if app.ReleaseState == "prerelease" {
		primary = append(primary, "This game is not released yet!")
	}
	if app.Type == "movie" {
		primary = append(primary, "This listing is for a movie")
	}

	if len(primary) > 0 {
		banners["primary"] = primary
	}

	// Wait
	wg.Wait()

	// Template
	t := appTemplate{}
	t.Fill(r, app.GetName())
	t.App = app
	t.Packages = packages
	t.Articles = news
	t.Banners = banners
	t.Prices = pricesString
	t.PricesCount = pricesCount
	t.Achievements = achievements
	t.Tags = tags
	t.DLC = dlc

	reviews, err := app.GetReviews()
	t.Reviews = reviews

	returnTemplate(w, r, "app", t)
}

type appTemplate struct {
	GlobalTemplate
	App          mysql.App
	Packages     []mysql.Package
	DLC          []mysql.App
	Articles     []appArticleTemplate
	Banners      map[string][]string
	Prices       string
	PricesCount  int
	Achievements []appAchievementTemplate
	Schema       steam.GameSchema
	Tags         []mysql.Tag
	Reviews      steam.ReviewsResponse
}

type appAchievementTemplate struct {
	Icon        string
	Name        string
	Description string
	Completed   float64
}

type appArticleTemplate struct {
	Title    string
	Contents template.HTML
	Author   string
}

func (a appAchievementTemplate) GetCompleted() float64 {
	return helpers.DollarsFloat(a.Completed)
}

type hcSeries struct {
	Type string
	Name string
	Data [][]int64
	Step bool
}
