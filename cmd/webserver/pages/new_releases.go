package pages

import (
	"net/http"
	"sync"
	"time"

	"github.com/gamedb/gamedb/cmd/webserver/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

func NewReleasesRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", newReleasesHandler)
	r.Get("/new-releases.json", newReleasesAjaxHandler)
	return r
}

func newReleasesHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	t := newReleasesTemplate{}
	t.fill(w, r, "New Releases", "")
	t.addAssetHighCharts()
	t.Days = config.Config.NewReleaseDays.GetInt()

	// Count apps
	{
		var filter = bson.D{
			{"release_date_unix", bson.M{"$lt": time.Now().Unix()}},
			{"release_date_unix", bson.M{"$gt": time.Now().AddDate(0, 0, -config.Config.NewReleaseDays.GetInt()).Unix()}},
		}

		t.Apps, err = mongo.CountDocuments(mongo.CollectionApps, filter, 86400)
		if err != nil {
			log.Err(err, r)
		}
	}

	//
	returnTemplate(w, r, "new_releases", t)
}

type newReleasesTemplate struct {
	GlobalTemplate
	Apps int64
	Days int
}

func newReleasesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, false)

	var wg sync.WaitGroup
	var count int64
	var filtered int64
	var apps []mongo.App
	var code = helpers.GetProductCC(r)
	var filter = bson.D{
		{"release_date_unix", bson.M{"$lt": time.Now().Unix()}},
		{"release_date_unix", bson.M{"$gt": time.Now().AddDate(0, 0, -config.Config.NewReleaseDays.GetInt()).Unix()}},
	}
	var countLock sync.Mutex

	wg.Add(1)
	go func() {

		defer wg.Done()

		var filter2 = filter

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
		apps, err = mongo.GetApps(query.GetOffset64(), 100, sort, filter2, projection, nil)
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
		count, err = mongo.CountDocuments(mongo.CollectionApps, filter, 60*60*24)
		countLock.Unlock()
		if err != nil {
			log.Err(err, r)
		}
	}()

	wg.Wait()

	//
	response := datatable.DataTablesResponse{}
	response.Output()
	response.RecordsTotal = count
	response.RecordsFiltered = filtered
	response.Draw = query.Draw

	for _, app := range apps {

		response.AddRow([]interface{}{
			app.ID,                          // 0
			app.GetName(),                   // 1
			app.GetIcon(),                   // 2
			app.GetPath(),                   // 3
			app.GetType(),                   // 4
			app.Prices.Get(code).GetFinal(), // 5
			helpers.GetAppReleaseDateNice(app.ReleaseDateUnix, app.ReleaseDate), // 6
			helpers.RoundFloatTo2DP(app.ReviewsScore),                           // 7
			app.PlayerPeakWeek, // 8
		})
	}

	returnJSON(w, r, response)
}
