package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi"
	slugify "github.com/gosimple/slug"
	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/steam"
)

func AppHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	slug := chi.URLParam(r, "slug")

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
	correctSLug := slugify.Make(app.GetName())
	if slug != correctSLug {
		http.Redirect(w, r, "/games/"+id+"/"+correctSLug, 302)
		return
	}

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
	wg.Add(1)
	go func() {

		// Get prices
		pricesResp, err := datastore.GetAppPrices(app.ID)
		if err != nil {
			logger.Error(err)
		}

		var prices [][]int64

		for _, v := range pricesResp {
			prices = append(prices, []int64{v.CreatedAt.Unix(), int64(v.PriceFinal)})
		}

		pricesBytes, err := json.Marshal(prices)
		if err != nil {
			logger.Error(err)
		}

		pricesString = string(pricesBytes)

		wg.Done()
	}()

	var news []datastore.Article
	wg.Add(1)
	go func() {

		// Get news
		news, err = datastore.GetArticles(idx, 1000)

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
	template := appTemplate{}
	template.Fill(r, app.GetName())
	template.App = app
	template.Packages = packages
	template.Articles = news
	template.Banners = banners
	template.Prices = pricesString
	template.Achievements = achievements
	template.Tags = tags

	returnTemplate(w, r, "app", template)
}

type appTemplate struct {
	GlobalTemplate
	App          mysql.App
	Packages     []mysql.Package
	Articles     []datastore.Article
	Banners      map[string][]string
	Prices       string
	Achievements []appAchievementTemplate
	Schema       steam.GameSchema
	Tags         []mysql.Tag
}

type appAchievementTemplate struct {
	Icon        string
	Name        string
	Description string
	Completed   float64
}

func (a appAchievementTemplate) GetCompleted() string {
	return helpers.DollarsFloat(a.Completed)
}

type hcSeries struct {
	Type string
	Name string
	Data [][]int64
	Step bool
}
