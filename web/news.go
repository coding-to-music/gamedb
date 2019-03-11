package web

import (
	"net/http"
	"sort"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/go-chi/chi"
)

func newsRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", newsHandler)
	r.Get("/news.json", newsAjaxHandler)
	return r
}

func newsHandler(w http.ResponseWriter, r *http.Request) {

	t := newsTemplate{}
	t.fill(w, r, "News", "All the news from all the games, all in one place.")

	apps, err := db.PopularApps()
	log.Err(err, r)

	if config.Config.IsLocal() {
		apps = apps[0:3]
	}

	client, ctx, err := db.GetDSClient()
	if err != nil {

		log.Err(err, r)
		return
	}

	for _, v := range apps {

		var news []db.News
		q := datastore.NewQuery(db.KindNews).Filter("app_id =", v.ID).Order("-date").Limit(3)
		_, err = client.GetAll(ctx, q, &news)
		err = db.HandleDSMultiError(err, db.OldNewsFields)
		log.Err(err, r)
		t.News = append(t.News, news...)
	}

	sort.Slice(t.News, func(i, j int) bool {
		return t.News[i].Date.Unix() > t.News[j].Date.Unix()
	})

	err = returnTemplate(w, r, "news", t)
	log.Err(err, r)
}

type newsTemplate struct {
	GlobalTemplate
	News []db.News
}

func newsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	setCacheHeaders(w, time.Hour*1)

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	var articles []db.News

	client, ctx, err := db.GetDSClient()
	if err != nil {

		log.Err(err, r)

	} else {

		q := datastore.NewQuery(db.KindNews).Order("-date").Limit(100).Offset(query.getOffset())
		_, err := client.GetAll(ctx, q, &articles)
		err = db.HandleDSMultiError(err, db.OldNewsFields)
		log.Err(err, r)

		for k, v := range articles {
			articles[k].Contents = helpers.BBCodeCompiler.Compile(v.Contents)
		}
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
