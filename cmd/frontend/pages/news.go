package pages

import (
	"net/http"
	"sync"
	"time"

	"github.com/gamedb/gamedb/cmd/frontend/pages/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
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
		zap.S().Error(err)
	}

	var appIDs []int
	for _, app := range apps {
		appIDs = append(appIDs, app.ID)
	}

	t.Articles, err = mongo.GetArticlesByAppIDs(appIDs, 0, time.Now().AddDate(0, 0, -7))
	if err != nil {
		zap.S().Error(err)
	}

	returnTemplate(w, r, "news", t)
}

type newsTemplate struct {
	globalTemplate
	Articles []mongo.Article
}

func newsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, true)

	var wg sync.WaitGroup

	var articles []elasticsearch.Article
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		var sorters = query.GetOrderElastic(map[string]string{
			"1": "time",
		})

		var search = query.GetSearchString("search")

		articles, filtered, err = elasticsearch.SearchArticles(query.GetOffset(), sorters, search)
		if err != nil {
			zap.S().Error(err)
		}
	}()

	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountDocuments(mongo.CollectionAppArticles, nil, 0)
		if err != nil {
			zap.S().Error(err)
		}
	}()

	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, count, filtered, nil)
	for _, article := range articles {

		response.AddRow([]interface{}{
			article.ID,               // 0
			article.Title,            // 1
			article.GetBody(),        // 2
			article.AppID,            // 3
			article.GetArticleIcon(), // 4
			article.Time,             // 5
			article.Score,            // 6
			article.GetAppName(),     // 7
			"",                       // 8
			article.GetAppPath(),     // 9
			article.GetDate(),        // 10
			article.TitleMarked,      // 11
		})
	}

	returnJSON(w, r, response)
}
