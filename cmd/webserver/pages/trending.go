package pages

import (
	"net/http"
	"strings"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
)

func TrendingRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", trendingHandler)
	r.Get("/trending.json", trendingAjaxHandler)
	r.Get("/charts.json", trendingChartsAjaxHandler)
	return r
}

func trendingHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	// Template
	t := trendingTemplate{}
	t.fill(w, r, "Trending", "")
	t.addAssetHighCharts()

	err = returnTemplate(w, r, "trending_apps", t)
	log.Err(err, r)
}

type trendingTemplate struct {
	GlobalTemplate
}

func trendingAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	query.limit(r)

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err, r)
		return
	}

	columns := map[string]string{
		"2": "player_peak_week",
		"3": "player_trend",
		"4": "player_trend",
	}

	gorm = gorm.Model(sql.App{})
	gorm = gorm.Select([]string{"id", "name", "icon", "prices", "player_trend", "player_peak_week"})
	gorm = gorm.Order(query.getOrderSQL(columns, helpers.GetCountryCode(r)))
	gorm = gorm.Limit(100)
	gorm = gorm.Offset(query.getOffset())

	var apps []sql.App
	gorm = gorm.Find(&apps)
	log.Err(gorm.Error, r)

	var code = helpers.GetCountryCode(r)

	count, err := sql.CountApps()
	log.Err(err)

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = int64(count)
	response.RecordsFiltered = int64(count)
	response.Draw = query.Draw
	response.limit(r)

	for _, app := range apps {
		response.AddRow([]interface{}{
			app.ID, // 0
			helpers.InsertNewLines(app.GetName(), 20), // 1
			app.GetIcon(),                          // 2
			app.GetPath(),                          // 3
			sql.GetPriceFormatted(app, code).Final, // 4
			app.PlayerTrend,                        // 5
			app.PlayerPeakWeek,                     // 6
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
	builder.SetFrom(helpers.InfluxGameDB, helpers.InfluxRetentionPolicyAllTime.String(), helpers.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW()-7d")
	builder.AddWhereRaw("(" + strings.Join(or, " OR ") + ")")
	builder.AddGroupByTime("1h")
	builder.AddGroupBy("app_id")
	builder.SetFillNone()

	resp, err := helpers.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	ret := map[string]helpers.HighChartsJson{}
	if len(resp.Results) > 0 {
		for _, v := range resp.Results[0].Series {
			ret[v.Tags["app_id"]] = helpers.InfluxResponseToHighCharts(v)
		}
	}

	err = returnJSON(w, r, ret)
	log.Err(err, r)
}
