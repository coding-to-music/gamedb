package pages

import (
	"net/http"
	"sync"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
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

	t.Apps, err = countNewReleaseApps()
	log.Err(err, r)

	returnTemplate(w, r, "new_releases", t)
}

type newReleasesTemplate struct {
	GlobalTemplate
	Apps int
	Days int
}

func newReleasesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

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

		var search = query.getSearchString("search")
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

		var projection = bson.M{"id": 1, "name": 1, "icon": 1, "type": 1, "prices": 1, "release_date_unix": 1, "release_date": 1, "player_peak_week": 1, "reviews_score": 1}
		var sort = query.getOrderMongo(columns)

		var err error
		apps, err = mongo.GetApps(query.getOffset64(), 100, sort, filter2, projection, nil)
		log.Err(err)

		countLock.Lock()
		filtered, err = mongo.CountDocuments(mongo.CollectionApps, filter2, 0)
		countLock.Unlock()
		log.Err(err)
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		countLock.Lock()
		count, err = mongo.CountDocuments(mongo.CollectionApps, filter, 60*60*24)
		countLock.Unlock()
		log.Err(err)
	}()

	wg.Wait()

	//
	response := DataTablesAjaxResponse{}
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

	response.output(w, r)
}

func countNewReleaseApps() (count int, err error) {

	var item = memcache.MemcacheNewReleaseAppsCount

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		var count int

		gorm, err := sql.GetMySQLClient()
		if err != nil {
			return count, err
		}

		gorm = gorm.Model(sql.App{})
		gorm = gorm.Where("release_date_unix < ?", time.Now().Unix())
		gorm = gorm.Where("release_date_unix > ?", time.Now().AddDate(0, 0, -config.Config.NewReleaseDays.GetInt()).Unix())
		gorm = gorm.Count(&count)

		return count, gorm.Error
	})

	return count, err
}
