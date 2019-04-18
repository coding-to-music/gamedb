package pages

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/website/pkg/helpers"
	"github.com/gamedb/website/pkg/influx"
	"github.com/gamedb/website/pkg/log"
	"github.com/gamedb/website/pkg/session"
	"github.com/gamedb/website/pkg/sql"
	"github.com/go-chi/chi"
)

func trendingRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", trendingHandler)
	r.Get("/trending.json", trendingAjaxHandler)
	r.Get("/charts.json", trendingChartsAjaxHandler)
	return r
}

func trendingHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*24)

	var err error

	// Template
	t := trendingTemplate{}
	t.fill(w, r, "Trending", "")
	t.addAssetHighCharts()

	t.Apps, err = countTrendingApps()
	log.Err(err, r)

	err = returnTemplate(w, r, "trending", t)
	log.Err(err, r)
}

type trendingTemplate struct {
	GlobalTemplate
	Apps int
}

func trendingAjaxHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{"draw", "order[0][column]", "order[0][dir]", "start"})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*1)

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

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
	gorm = gorm.Where("player_trend >= 100 OR player_trend <= -100")
	gorm = gorm.Order(query.getOrderSQL(columns, session.GetCountryCode(r)))
	gorm = gorm.Limit(50)
	gorm = gorm.Offset(query.getOffset())

	var apps []sql.App
	gorm = gorm.Find(&apps)
	log.Err(gorm.Error, r)

	var code = session.GetCountryCode(r)

	count, err := countTrendingApps()
	log.Err(err)

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(count)
	response.RecordsFiltered = strconv.Itoa(count)
	response.Draw = query.Draw

	for _, app := range apps {
		response.AddRow([]interface{}{
			app.ID,                                 // 0
			app.GetName(),                          // 1
			app.GetIcon(),                          // 2
			app.GetPath(),                          // 3
			sql.GetPriceFormatted(app, code).Final, // 4
			app.PlayerTrend,                        // 5
			app.PlayerPeakWeek,                     // 6
		})
	}

	response.output(w, r)
}

func countTrendingApps() (count int, err error) {

	var item = helpers.MemcacheTrendingAppsCount

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		var count int

		gorm, err := sql.GetMySQLClient()
		if err != nil {
			return count, err
		}

		gorm = gorm.Model(sql.App{})
		gorm = gorm.Where("player_trend >= 100 OR player_trend <= -100")
		gorm = gorm.Count(&count)

		return count, gorm.Error
	})

	return count, err
}

func trendingChartsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	retu := setAllowedQueries(w, r, []string{"ids"})
	if retu {
		return
	}

	setCacheHeaders(w, time.Hour/2)

	idsString := r.URL.Query().Get("ids")
	idsSlice := strings.Split(idsString, ",")

	if len(idsSlice) == 0 {
		return
	}

	if len(idsSlice) > 50 {
		idsSlice = idsSlice[0:50]
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
	builder.SetFrom("GameDB", "alltime", "apps")
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

	ret := map[string]influx.HighChartsJson{}
	if len(resp.Results) > 0 {
		for _, v := range resp.Results[0].Series {
			ret[v.Tags["app_id"]] = influx.InfluxResponseToHighCharts(v)
		}
	}

	b, err := json.Marshal(ret)
	if err != nil {
		log.Err(err, r)
		return
	}

	err = returnJSON(w, r, b)
	if err != nil {
		log.Err(err, r)
		return
	}
}
