package pages

import (
	"net/http"
	"sync"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/bson"
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

	feeds, err := mongo.GetAppArticlesGroupedByFeed()
	if err != nil {
		log.ErrS(err)
	}

	t.Feeds = feeds

	returnTemplate(w, r, t)
}

type newsTemplate struct {
	globalTemplate
	Feeds []mongo.ArticleFeed
}

func newsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, false)

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

		var filters []elastic.Query

		// Search
		var search = query.GetSearchString("search")

		// Filter
		switch query.GetSearchString("filter") {
		case "mine":

			playerID := session.GetPlayerIDFromSesion(r)
			if playerID == 0 {
				break
			}

			apps, err := mongo.GetPlayerAppsByPlayer(playerID, 0, 0, nil, bson.M{"_id": 1}, bson.D{{"app_time", bson.M{"$gt": 0}}})
			if err != nil {
				log.ErrS(err)
				break
			}

			var appIDs []interface{}
			for _, v := range apps {
				appIDs = append(appIDs, v.AppID)
			}

			log.InfoS(len(appIDs))

			filters = append(filters, elastic.NewTermsQuery("app_id", appIDs...))

		case "popular":

			apps, err := mongo.PopularApps()
			if err != nil {
				log.ErrS(err)
				break
			}

			var appIDs []interface{}
			for _, v := range apps {
				appIDs = append(appIDs, v.ID)
			}

			log.InfoS(len(appIDs))

			filters = append(filters, elastic.NewTermsQuery("app_id", appIDs...))
		}

		// Feed
		var feed = query.GetSearchString("feed")
		if feed != "" {
			filters = append(filters, elastic.NewTermQuery("feed", feed))
		}

		//
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
			article.GetFeedName(),    // 11
		})
	}

	returnJSON(w, r, response)
}
