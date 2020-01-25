package pages

import (
	"net/http"
	"sync"
	"time"

	"github.com/gamedb/gamedb/cmd/webserver/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
)

func NewsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", newsHandler)
	r.Get("/news.json", newsAjaxHandler)
	return r
}

func newsHandler(w http.ResponseWriter, r *http.Request) {

	t := newsTemplate{}
	t.fill(w, r, "News", "All the news from all the games")

	apps, err := mongo.PopularApps()
	if err != nil {
		log.Err(err, r)
	}

	var appIDs []int
	for _, app := range apps {
		appIDs = append(appIDs, app.ID)
	}

	t.Articles, err = mongo.GetArticlesByApps(appIDs, 0, time.Now().AddDate(0, 0, -7))
	if err != nil {
		log.Err(err, r)
	}

	count, err := mongo.CountDocuments(mongo.CollectionAppArticles, nil, 0)
	if err != nil {
		log.Err(err, r)
	}

	t.Count = helpers.ShortHandNumber(count)

	returnTemplate(w, r, "news", t)
}

type newsTemplate struct {
	GlobalTemplate
	Articles []mongo.Article
	Count    string
}

func newsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, true)

	var wg sync.WaitGroup

	var articles []mongo.Article
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		articles, err = mongo.GetArticles(query.GetOffset64())
		log.Err(err, r)

		for k, v := range articles {
			articles[k].Contents = helpers.BBCodeCompiler.Compile(v.Contents)
		}
	}()

	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountDocuments(mongo.CollectionAppArticles, nil, 0)
		log.Err(err, r)
	}()

	wg.Wait()

	response := datatable.DataTablesResponse{}
	response.Output()
	response.RecordsTotal = count
	response.RecordsFiltered = count
	response.Draw = query.Draw
	response.Limit(r)

	for _, v := range articles {
		response.AddRow(v.OutputForJSON())
	}

	returnJSON(w, r, response)
}
