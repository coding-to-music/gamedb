package pages

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
	"github.com/olivere/elastic/v7"
)

func releaseDatesRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", releaseDatesHandler)
	r.Get("/release-dates.json", releaseDatesAjaxHandler)
	return r
}

func releaseDatesHandler(w http.ResponseWriter, r *http.Request) {

	t := releaseDatesTemplate{}
	t.fill(w, r, "apps_release_dates", "Release Dates", "Games with unconventional release dates")

	returnTemplate(w, r, t)
}

type releaseDatesTemplate struct {
	globalTemplate
}

func releaseDatesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, false)

	var wg sync.WaitGroup

	var apps []elasticsearch.App
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		columns := map[string]string{
			"2": "followers, name.raw asc",
			"4": "_score",
		}

		// Boost the bad ones and show in reverse order
		boolQuery := elastic.NewBoolQuery().
			Filter(
				elastic.NewTermQuery("release_date", 0),
			).
			Should(
				// match query is case insensitive
				elastic.NewMatchQuery("release_date_original", strings.Join([]string{"tbd", "tba", "\"coming soon\"", "fall", "winter", "spring", "autumn", "q1", "q2", "q3", "q4"}, " OR ")),
				elastic.NewRegexpQuery("release_date_original", "[0-9]{4}"),
			).
			MustNot(
				elastic.NewTermQuery("release_date_original.raw", ""),
			)

		apps, filtered, err = elasticsearch.SearchAppsAdvanced(query.GetOffset(), 100, "", query.GetOrderElastic(columns), boolQuery)
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Wait()

	//
	var response = datatable.NewDataTablesResponse(r, query, filtered, filtered, nil)
	for _, app := range apps {

		response.AddRow([]interface{}{
			app.ID,                          // 0
			app.GetName(),                   // 1
			app.GetIcon(),                   // 2
			app.GetPath(),                   // 3
			app.GetFollowers(),              // 4
			helpers.GetAppStoreLink(app.ID), // 5
			app.ReleaseDateOriginal,         // 6
			app.Score,                       // 7
		})
	}

	returnJSON(w, r, response)
}
