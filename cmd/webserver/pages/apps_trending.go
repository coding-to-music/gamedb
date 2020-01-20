package pages

import (
	"net/http"
	"strings"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
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
	t.fill(w, r, "Trending", "")
	t.addAssetHighCharts()

	returnTemplate(w, r, "trending_apps", t)
}

type trendingTemplate struct {
	GlobalTemplate
}

func trendingAppsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	if err != nil {
		log.Err(err, r)
	}

	query.limit(r)

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err, r)
		return
	}

	gorm = gorm.Model(sql.App{})
	gorm = gorm.Select([]string{"id", "name", "icon", "prices", "player_trend", "player_peak_week"})
	gorm = gorm.Limit(100)

	columns := map[string]string{
		"2": "player_peak_week",
		"3": "player_trend",
		"4": "player_trend",
	}
	gorm = query.setOrderOffsetGorm(gorm, columns, "3")

	var apps []sql.App
	gorm = gorm.Find(&apps)
	if gorm.Error != nil {
		log.Err(gorm.Error, r)
	}

	count, err := mongo.CountDocuments(mongo.CollectionApps, nil, 0)
	if err != nil {
		log.Err(err, r)
	}

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = count
	response.RecordsFiltered = count
	response.Draw = query.Draw
	response.limit(r)

	var code = helpers.GetProductCC(r)

	for _, app := range apps {

		response.AddRow([]interface{}{
			app.ID,                              // 0
			app.GetName(),                       // 1
			app.GetIcon(),                       // 2
			app.GetPath(),                       // 3
			app.GetPrice(code).GetFinal(),       // 4
			helpers.TrendValue(app.PlayerTrend), // 5
			app.PlayerPeakWeek,                  // 6
		})
	}

	response.output(w, r)
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
