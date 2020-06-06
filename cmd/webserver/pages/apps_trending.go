package pages

import (
	"net/http"
	"strings"
	"sync"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

func trendingRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", trendingHandler)
	r.Get("/trending.json", trendingAppsAjaxHandler)
	r.Get("/charts.json", trendingChartsAjaxHandler)
	return r
}

func trendingHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := trendingTemplate{}
	t.fill(w, r, "Trending", "Trending Steam Games")
	t.addAssetHighCharts()

	returnTemplate(w, r, "trending_apps", t)
}

type trendingTemplate struct {
	GlobalTemplate
}

func trendingAppsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, true)

	var filter = bson.D{}
	var search = query.GetSearchString("search")
	if search != "" {
		filter = bson.D{{Key: "$text", Value: bson.M{"$search": search}}}
	}

	var wg sync.WaitGroup
	var countLock sync.Mutex

	// Count
	var apps []mongo.App
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		columns := map[string]string{
			"2": "player_peak_week",
			"3": "player_trend",
			"4": "player_trend",
		}

		projection := bson.M{"_id": 1, "name": 1, "icon": 1, "prices": 1, "player_trend": 1, "player_peak_week": 1}
		order := query.GetOrderMongo(columns)
		offset := query.GetOffset64()

		apps, err = mongo.GetApps(offset, 100, order, filter, projection)
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Get filtered count
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		countLock.Lock()
		filtered, err = mongo.CountDocuments(mongo.CollectionApps, filter, 0)
		countLock.Unlock()
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Get count
	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		countLock.Lock()
		count, err = mongo.CountDocuments(mongo.CollectionApps, nil, 0)
		countLock.Unlock()
		if err != nil {
			log.Err(err, r)
		}
	}()

	wg.Wait()

	var code = session.GetProductCC(r)
	var response = datatable.NewDataTablesResponse(r, query, count, filtered, nil)
	for _, app := range apps {

		response.AddRow([]interface{}{
			app.ID,                              // 0
			app.GetName(),                       // 1
			app.GetIcon(),                       // 2
			app.GetPath(),                       // 3
			app.Prices.Get(code).GetFinal(),     // 4
			helpers.TrendValue(app.PlayerTrend), // 5
			app.PlayerPeakWeek,                  // 6
		})
	}

	returnJSON(w, r, response)
}

func trendingChartsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	idsString := r.URL.Query().Get("ids")
	idsSlice := strings.Split(idsString, ",")

	if len(idsSlice) == 0 {
		return
	}

	if len(idsSlice) > 100 {
		idsSlice = idsSlice[0:100]
	}

	var or []string
	for _, v := range idsSlice {
		v = strings.TrimSpace(v)
		if v != "" {
			or = append(or, `"app_id" = '`+v+`'`)
		}
	}

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW()-7d")
	builder.AddWhereRaw("(" + strings.Join(or, " OR ") + ")")
	builder.AddGroupByTime("1h")
	builder.AddGroupBy("app_id")
	builder.SetFillNone()

	resp, err := influx.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	ret := map[string]influx.HighChartsJSON{}
	if len(resp.Results) > 0 {
		for _, v := range resp.Results[0].Series {
			ret[v.Tags["app_id"]] = influx.InfluxResponseToHighCharts(v)
		}
	}

	returnJSON(w, r, ret)
}
