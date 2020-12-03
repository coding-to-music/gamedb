package pages

import (
	"net/http"
	"sync"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
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
	t.fill(w, r, "news", "News", "All the news from all the games on Steam")

	feeds, err := elasticsearch.AggregateArticleFeeds()
	if err != nil {
		log.ErrS(err)
	}

	t.Feeds = feeds

	returnTemplate(w, r, t)
}

type newsTemplate struct {
	globalTemplate
	Feeds []helpers.TupleStringInt
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
		var filters []elastic.Query

		switch query.GetSearchString("filter") {
		case "mine":
			filters = append(filters, elastic.NewTermQuery("", ""))
		case "popular":
			filters = append(filters, elastic.NewTermQuery("", ""))
		}

		var feed = query.GetSearchString("feed")
		if feed != "" {
			filters = append(filters, elastic.NewTermQuery("feed", feed))
		}

		articles, filtered, err = elasticsearch.SearchArticles(query.GetOffset(), sorters, search, filters)
		if err != nil {
			log.ErrS(err)
		}
	}()

	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountDocuments(mongo.CollectionAppArticles, nil, 0)
		if err != nil {
			log.ErrS(err)
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
			article.GetAppPath(),     // 8
			article.GetDate(),        // 9
			article.TitleMarked,      // 10
			article.FeedName,         // 11
		})
	}

	returnJSON(w, r, response)
}
