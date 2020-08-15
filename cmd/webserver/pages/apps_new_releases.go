package pages

import (
	"html/template"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

func newReleasesRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", newReleasesHandler)
	r.Get("/new-releases.json", newReleasesAjaxHandler)
	return r
}

func newReleasesHandler(w http.ResponseWriter, r *http.Request) {

	var days = config.Config.NewReleaseDays.Get()

	t := newReleasesTemplate{}
	t.fill(w, r, "New Releases", template.HTML("Newly released games in the last "+days+" days"))
	t.addAssetHighCharts()
	t.Days = days

	//
	returnTemplate(w, r, "new_releases", t)
}

type newReleasesTemplate struct {
	globalTemplate
	Days string
}

func newReleasesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	var query = datatable.NewDataTableQuery(r, false)
	var code = session.GetProductCC(r)
	var countLock sync.Mutex
	var wg sync.WaitGroup

	var apps []mongo.App
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var days = config.Config.NewReleaseDays.GetInt()

		i, err := strconv.Atoi(query.GetSearchString("days"))
		if err == nil && i >= 1 && i <= 30 {
			days = i
		}

		var filter2 = bson.D{
			{"release_date_unix", bson.M{"$lt": time.Now().Unix()}},
			{"release_date_unix", bson.M{"$gt": time.Now().AddDate(0, 0, -days).Unix()}},
		}

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

		apps, err = mongo.GetApps(query.GetOffset64(), 100, sort, filter2, projection)
		if err != nil {
			zap.S().Error(err)
		}

		countLock.Lock()
		filtered, err = mongo.CountDocuments(mongo.CollectionApps, filter2, 0)
		countLock.Unlock()
		if err != nil {
			zap.S().Error(err)
		}
	}()

	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		var filter = bson.D{
			{"release_date_unix", bson.M{"$lt": time.Now().Unix()}},
			{"release_date_unix", bson.M{"$gt": time.Now().AddDate(0, 0, -config.Config.NewReleaseDays.GetInt()).Unix()}},
		}

		countLock.Lock()
		count, err = mongo.CountDocuments(mongo.CollectionApps, filter, 60*60*24)
		countLock.Unlock()
		if err != nil {
			zap.S().Error(err)
		}
	}()

	wg.Wait()

	//
	var response = datatable.NewDataTablesResponse(r, query, count, filtered, nil)
	for _, app := range apps {

		var release = helpers.GetAppReleaseDateNice(app.ReleaseDateOriginal, app.ReleaseDateUnix, app.ReleaseDate)
		var score = helpers.RoundFloatTo2DP(app.ReviewsScore)
		var price = app.Prices.Get(code).GetFinal()

		response.AddRow([]interface{}{
			app.ID,             // 0
			app.GetName(),      // 1
			app.GetIcon(),      // 2
			app.GetPath(),      // 3
			app.GetType(),      // 4
			price,              // 5
			release,            // 6
			score,              // 7
			app.PlayerPeakWeek, // 8
		})
	}

	returnJSON(w, r, response)
}
