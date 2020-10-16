package pages

import (
	"net/http"
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

func (t releaseDatesTemplate) includes() []string {
	return []string{"includes/apps_header.gohtml"}
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
			"0": "name.raw",
			"2": "followers, name.raw asc",
		}

		boolQuery := elastic.NewBoolQuery().Filter(
			elastic.NewTermQuery("release_date", 0),
		).MustNot(
			elastic.NewTermQuery("release_date_original.raw", ""),
			// elastic.NewRegexpQuery("release_date_original", "/(tbd|tba|coming soon|^$)/i"),
		)

		apps, filtered, err = elasticsearch.SearchAppsAdvanced(query.GetOffset(), 1000, query.GetSearchString("search"), query.GetOrderElastic(columns), boolQuery)
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
		})
	}

	returnJSON(w, r, response)
}
