package pages

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gamedb/gamedb/pkg/helpers/search"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
	"github.com/olivere/elastic/v7"
)

func SearchRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", searchHandler)
	// r.Get("/sales.json", searchAjaxHandler)
	return r
}

func searchHandler(w http.ResponseWriter, r *http.Request) {

	t := searchTemplate{}
	t.fill(w, r, "Search", "Search all of Game DB")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {

		defer wg.Done()

		// var err error
		// t.HighestOrder, err = mongo.GetHighestSaleOrder()
		// log.Err(err, r)
	}()

	// Get categories
	wg.Add(1)
	go func() {

		defer wg.Done()

		client, ctx, err := search.GetElastic()
		if err != nil {
			log.Err(err)
			return
		}

		// Search with a term query
		termQuery := elastic.NewTermQuery("jam", "olivere")
		searchResult, err := client.Search().
			Index("gdb-search").
			Query(termQuery).
			// Sort("user", true).
			From(0).
			Size(10).
			// Pretty(true).
			Human(true).
			Do(ctx)

		if err != nil {
			log.Err(err)
			return
		}

		for _, hit := range searchResult.Hits.Hits {

			// Deserialize hit.Source into a Tweet (could also be just a map[string]interface{}).
			var result search.SearchResult
			err := json.Unmarshal(hit.Source, &result)
			if err != nil {
				log.Err(err)
			}

			t.Results = append(t.Results, result)
		}
	}()

	// Wait
	wg.Wait()

	returnTemplate(w, r, "search", t)
}

type searchTemplate struct {
	GlobalTemplate
	Results []search.SearchResult
}
