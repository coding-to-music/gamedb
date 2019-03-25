package web

import (
	"net/http"
	"time"

	"github.com/gamedb/website/config"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/mongo"
	"github.com/gamedb/website/sql"
	"github.com/go-chi/chi"
)

func newsRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", newsHandler)
	r.Get("/news.json", newsAjaxHandler)
	return r
}

func newsHandler(w http.ResponseWriter, r *http.Request) {

	setCacheHeaders(w, time.Hour)

	t := newsTemplate{}
	t.fill(w, r, "News", "All the news from all the games, all in one place.")

	apps, err := sql.PopularApps()
	log.Err(err, r)

	if config.Config.IsLocal() && len(apps) >= 3 {
		apps = apps[0:3]
	}

	var appIDs []int
	for _, v := range apps {
		appIDs = append(appIDs, v.ID)
	}

	t.Articles, err = mongo.GetArticlesByAppIDs(appIDs)
	log.Err(err, r)

	err = returnTemplate(w, r, "news", t)
	log.Err(err, r)
}

type newsTemplate struct {
	GlobalTemplate
	Articles []mongo.Article
}

func newsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	setCacheHeaders(w, time.Hour*1)

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	articles, err := mongo.GetArticles(query.getOffset64())

	for k, v := range articles {
		articles[k].Contents = helpers.BBCodeCompiler.Compile(v.Contents)
	}

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = "10000"
	response.RecordsFiltered = "10000"
	response.Draw = query.Draw

	for _, v := range articles {
		response.AddRow(v.OutputForJSON())
	}

	response.output(w, r)
}
