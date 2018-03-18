package web

import (
	"net/http"
	"strconv"

	"github.com/Jleagle/go-helpers/logger"
	"github.com/go-chi/chi"
	slugify "github.com/gosimple/slug"
	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/mysql"
)

func AppsHandler(w http.ResponseWriter, r *http.Request) {

	// Get apps
	apps, err := mysql.SearchApps(r.URL.Query(), 96, "id DESC", []string{})
	if err != nil {
		logger.Error(err)
	}

	// Get apps count
	count, err := mysql.CountApps()
	if err != nil {
		logger.Error(err)
	}

	// Template
	template := appsTemplate{}
	template.Fill(r)
	template.Apps = apps
	template.Count = count

	returnTemplate(w, r, "apps", template)
}

type appsTemplate struct {
	GlobalTemplate
	Apps  []mysql.App
	Count int
}

func AppHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	slug := chi.URLParam(r, "slug")

	idx, err := strconv.Atoi(id)
	if err != nil {
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

	// Get prices
	prices, err := datastore.GetAppPrices(app.ID)
	if err != nil {
		logger.Error(err)
	}

	// Redirect to correct slug
	correctSLug := slugify.Make(app.Name)
	if slug != "" && app.Name != "" && slug != correctSLug {
		http.Redirect(w, r, "/apps/"+id+"/"+correctSLug, 302)
		return
	}

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

	// Get news
	news, err := datastore.GetArticles(idx, 1000)

	// Get packages
	packages, err := mysql.GetPackagesAppIsIn(app.ID)
	if err != nil {
		logger.Error(err)
	}

	// Template
	template := appTemplate{}
	template.Fill(r)
	template.App = app
	template.Packages = packages
	template.Articles = news
	template.Banners = banners
	template.Prices = prices

	returnTemplate(w, r, "app", template)
}

type appTemplate struct {
	GlobalTemplate
	App      mysql.App
	Packages []mysql.Package
	Articles []datastore.Article
	Banners  map[string][]string
	Prices   []datastore.AppPrice
}
