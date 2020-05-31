package pages

import (
	"html/template"
	"net/http"
	"sync"
	"time"

	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

func newReleasesRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", newReleasesHandler)
	r.Get("/new-releases.json", newReleasesAjaxHandler)
	return r
}

func newReleasesHandler(w http.ResponseWriter, r *http.Request) {

	t := newReleasesTemplate{}
	t.fill(w, r, "New Releases", template.HTML("Games released in the last "+config.Config.NewReleaseDays.Get()+" days"))
	t.addAssetHighCharts()

	//
	returnTemplate(w, r, "new_releases", t)
}

type newReleasesTemplate struct {
	GlobalTemplate
}

var newReleasesFilter = bson.D{
	{"release_date_unix", bson.M{"$lt": time.Now().Unix()}},
	{"release_date_unix", bson.M{"$gt": time.Now().AddDate(0, 0, -config.Config.NewReleaseDays.GetInt()).Unix()}},
}

func newReleasesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, false)

	var wg sync.WaitGroup
	var count int64
	var filtered int64
	var apps []mongo.App
	var code = session.GetProductCC(r)
	var countLock sync.Mutex

	wg.Add(1)
	go func() {

		defer wg.Done()

		var filter2 = newReleasesFilter

		var search = query.GetSearchString("search")
		if search != "" {
			filter2 = append(filter2, bson.E{Key: "$text", Value: bson.M{"$search": search}})
		}

		var columns = map[string]string{
			"1": "prices." + string(code) + ".final",
			"2": "reviews_score",
			"3": "player_peak_week",
			"4": "release_date_unix",
			"5": "player_trend",
		}

		var projection = bson.M{"_id": 1, "name": 1, "icon": 1, "type": 1, "prices": 1, "release_date_unix": 1, "release_date": 1, "player_peak_week": 1, "reviews_score": 1}
		var sort = query.GetOrderMongo(columns)

		var err error
		apps, err = mongo.GetApps(query.GetOffset64(), 100, sort, filter2, projection)
		if err != nil {
			log.Err(err, r)
		}

		countLock.Lock()
		filtered, err = mongo.CountDocuments(mongo.CollectionApps, filter2, 0)
		countLock.Unlock()
		if err != nil {
			log.Err(err, r)
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		countLock.Lock()
		count, err = mongo.CountDocuments(mongo.CollectionApps, newReleasesFilter, 60*60*24)
		countLock.Unlock()
		if err != nil {
			log.Err(err, r)
		}
	}()

	wg.Wait()

	//
	var response = datatable.NewDataTablesResponse(r, query, count, filtered)
	for _, app := range apps {

		var release = helpers.GetAppReleaseDateNice(app.ReleaseDateOriginal, app.ReleaseDateUnix, app.ReleaseDate)

		response.AddRow([]interface{}{
			app.ID,                          // 0
			app.GetName(),                   // 1
			app.GetIcon(),                   // 2
			app.GetPath(),                   // 3
			app.GetType(),                   // 4
			app.Prices.Get(code).GetFinal(), // 5
			release,                         // 6
			helpers.RoundFloatTo2DP(app.ReviewsScore), // 7
			app.PlayerPeakWeek,                        // 8
		})
	}

	returnJSON(w, r, response)
}
