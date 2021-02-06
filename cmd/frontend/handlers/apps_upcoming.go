package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/bson"
)

const upcomingFilterHours = time.Hour * -12

var upcomingFilter = bson.D{{"release_date_unix", bson.M{"$gte": time.Now().Add(upcomingFilterHours).Unix()}}}

func upcomingRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", upcomingHandler)
	r.Get("/upcoming.json", upcomingAjaxHandler)
	return r
}

func upcomingHandler(w http.ResponseWriter, r *http.Request) {

	t := upcomingTemplate{}
	t.fill(w, r, "upcoming", "Upcoming", "Games with a release date in the future")

	returnTemplate(w, r, t)
}

type upcomingTemplate struct {
	globalTemplate
}

func upcomingAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, false)

	var wg sync.WaitGroup

	var apps []elasticsearch.App
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		columns := map[string]string{
			"1": "followers, name asc",
			"2": "release_date_rounded, followers desc, name asc",
		}

		var filters = []elastic.Query{
			elastic.NewRangeQuery("release_date").From(time.Now().Add(upcomingFilterHours).Unix()),
		}

		apps, filtered, err = elasticsearch.SearchAppsAdvanced(query.GetOffset(), 100, query.GetSearchString("search"), query.GetOrderElastic(columns), elastic.NewBoolQuery().Filter(filters...))
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get count
	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountDocuments(mongo.CollectionApps, upcomingFilter, 60*60)
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Wait()

	//
	var response = datatable.NewDataTablesResponse(r, query, count, filtered, nil)
	for _, app := range apps {

		date := time.Unix(app.ReleaseDate, 0).Format(helpers.DateYear)

		response.AddRow([]interface{}{
			app.ID,                          // 0
			app.GetName(),                   // 1
			app.GetIcon(),                   // 2
			app.GetPath(),                   // 3
			app.GetType(),                   // 4
			"",                              // 5
			app.GetReleaseDateNiceRounded(), // 6
			app.GetFollowers(),              // 7
			helpers.GetAppStoreLink(app.ID), // 8
			app.ReleaseDate,                 // 9
			date,                            // 10
		})
	}

	returnJSON(w, r, response)
}
