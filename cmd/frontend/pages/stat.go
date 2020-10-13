package pages

import (
	"html/template"
	"net/http"
	"strconv"
	"sync"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

func StatRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", statHandler)
	r.Get("/time.json", statTimeAjaxHandler)
	r.Get("/apps.json", statAppsAjaxHandler)
	r.Get("/{slug}", statHandler)

	return r
}

func statHandler(w http.ResponseWriter, r *http.Request) {

	typex := statPathToConst(chi.URLParam(r, "type"))

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid App ID"})
		return
	}

	stat, err := mongo.GetStat(typex, id)
	if err == mongo.ErrNoDocuments {
		returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Unable to find " + typex.Title()})
		return
	} else if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: err.Error()})
		return
	}

	t := statTagsTemplate{}
	t.fill(w, r, "stat", stat.Name+" "+stat.Type.Title(), template.HTML(stat.Name+" "+stat.Type.Title()))
	t.addAssetHighCharts()

	t.Stat = stat

	returnTemplate(w, r, t)
}

type statTagsTemplate struct {
	globalTemplate
	Stat mongo.Stat
}

func statAppsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	idx, err := strconv.Atoi(id)
	if err != nil {
		return
	}

	query := datatable.NewDataTableQuery(r, false)

	var wg sync.WaitGroup
	var filter = bson.D{{Key: statPathToConst(chi.URLParam(r, "type")).MongoCol(), Value: idx}}
	var countLock sync.Mutex

	var apps []mongo.App
	wg.Add(1)
	go func() {

		defer wg.Done()

		var code = session.GetProductCC(r)

		var columns = map[string]string{
			"1": "player_peak_week",
			"2": "prices." + string(code) + ".final",
			"3": "reviews_score",
		}

		var projection = bson.M{
			"_id":              1,
			"name":             1,
			"icon":             1,
			"player_peak_week": 1,
			"prices":           1,
			"reviews_score":    1,
		}
		var sort = query.GetOrderMongo(columns)

		var err error
		apps, err = mongo.GetApps(query.GetOffset64(), 100, sort, filter, projection)
		if err != nil {
			log.ErrS(err)
		}
	}()

	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		countLock.Lock()
		count, err = mongo.CountDocuments(mongo.CollectionApps, filter, 60*60*24)
		countLock.Unlock()
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Wait()

	//
	var code = session.GetProductCC(r)
	var response = datatable.NewDataTablesResponse(r, query, count, count, nil)
	for _, app := range apps {

		for k, v := range app.Achievements {
			app.Achievements[k].Key = helpers.GetAchievementIcon(app.ID, v.Key)
		}

		response.AddRow([]interface{}{
			app.ID,                          // 0
			app.GetName(),                   // 1
			app.GetIcon(),                   // 2
			app.GetPath(),                   // 3
			app.PlayerPeakWeek,              // 4
			app.Prices.Get(code).GetFinal(), // 5
			app.GetReviewScore(),            // 6
			app.GetStoreLink(),              // 7
		})
	}

	returnJSON(w, r, response)
}

func statTimeAjaxHandler(w http.ResponseWriter, r *http.Request) {

	idx, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return
	}

	key := mongo.Stat{
		Type: statPathToConst(chi.URLParam(r, "type")),
		ID:   idx,
	}.GetKey()

	hc := influx.HighChartsJSON{}
	code := session.GetProductCC(r)

	callback := func() (interface{}, error) {

		builder := influxql.NewBuilder()
		builder.AddSelect(`max("apps_count")`, "max_apps_count")
		builder.AddSelect(`max("apps_percent")`, "max_apps_percent")
		builder.AddSelect(`max("mean_score")`, "max_mean_score")
		builder.AddSelect(`max("mean_players")`, "max_mean_players")
		builder.AddSelect(`max("mean_price_`+string(code)+`")`, "max_mean_price_"+string(code))
		builder.AddSelect(`max("median_score")`, "max_median_score")
		builder.AddSelect(`max("median_players")`, "max_median_players")
		builder.AddSelect(`max("median_price_`+string(code)+`")`, "max_median_price_"+string(code))
		builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementStats.String())
		builder.AddWhere("key", "=", key)
		builder.AddWhere("time", ">", "now()-365d")
		builder.AddGroupByTime("1d")
		builder.SetFillNone()

		resp, err := influx.InfluxQuery(builder)
		if err != nil {
			log.Err(err.Error(), zap.String("query", builder.String()))
			return hc, err
		}

		if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

			hc = influx.InfluxResponseToHighCharts(resp.Results[0].Series[0], true, influx.FilterAtLeastOne)
		}

		return hc, err
	}

	var item = memcache.MemcacheStatTime(key, code)
	err = memcache.GetSetInterface(item.Key, item.Expiration, &hc, callback)
	if err != nil {
		log.ErrS(err)
	}

	returnJSON(w, r, hc)
}
