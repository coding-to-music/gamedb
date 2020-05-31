package pages

import (
	"net/http"
	"sync"
	"time"

	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	elasticHelpers "github.com/gamedb/gamedb/pkg/elastic"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"github.com/olivere/elastic/v7"
)

func NewsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", newsHandler)
	r.Get("/news.json", newsAjaxHandler)
	return r
}

func newsHandler(w http.ResponseWriter, r *http.Request) {

	t := newsTemplate{}
	t.fill(w, r, "News", "All the news from all the games on Steam")

	apps, err := mongo.PopularApps()
	if err != nil {
		log.Err(err, r)
	}
	t.Apps = apps

	var appIDs []int
	for _, app := range apps {
		appIDs = append(appIDs, app.ID)
	}

	t.Articles, err = mongo.GetArticlesByAppIDs(appIDs, 0, time.Now().AddDate(0, 0, -7))
	if err != nil {
		log.Err(err, r)
	}

	returnTemplate(w, r, "news", t)
}

type newsTemplate struct {
	GlobalTemplate
	Articles []mongo.Article
	Apps     []mongo.App
}

func newsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, true)

	var wg sync.WaitGroup

	var articles []elasticHelpers.Article
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		var search = query.GetSearchString("search")
		var sorters = query.GetOrderElastic(map[string]string{
			"1": "time",
		})

		var q elastic.Query
		if search != "" {
			q = elastic.NewBoolQuery().
				Must(elastic.NewMatchQuery("title", search)).
				Should(elastic.NewTermQuery("title", search).Boost(5))
		}

		articles, filtered, err = elasticHelpers.SearchArticles(100, query.GetOffset(), q, sorters)
		if err != nil {
			log.Err(err, r)
		}
	}()

	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountDocuments(mongo.CollectionAppArticles, nil, 0)
		if err != nil {
			log.Err(err, r)
		}
	}()

	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, count, filtered)
	for _, v := range articles {

		var appIcon = helpers.GetAppIcon(v.AppID, v.AppIcon)
		var appPath = helpers.GetAppPath(v.AppID, v.AppName) + "#news"
		var appName = helpers.GetAppName(v.AppID, v.AppName)
		var date = time.Unix(v.Time, 0).Format(helpers.DateYearTime)

		response.AddRow([]interface{}{
			v.ID,        // 0
			v.Title,     // 1
			v.GetBody(), // 2
			v.AppID,     // 3
			v.AppIcon,   // 4
			v.Time,      // 5
			v.Score,     // 6
			appName,     // 7
			appIcon,     // 8
			appPath,     // 9
			date,        // 10
		})
	}

	returnJSON(w, r, response)
}
