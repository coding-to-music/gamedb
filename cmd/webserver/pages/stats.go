package pages

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/cmd/webserver/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
)

func StatsRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", statsHandler)
	r.Get("/client-players.json", statsClientPlayersHandler)
	r.Get("/release-dates.json", statsDatesHandler)
	r.Get("/app-scores.json", statsScoresHandler)
	r.Get("/app-types.json", statsTypesHandler)
	return r
}

func statsHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*24)

	// Template
	t := statsTemplate{}
	t.fill(w, r, "Stats", "Some interesting Steam Store stats.")
	t.addAssetHighCharts()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PlayersCount, err = mongo.CountPlayers()
		log.Err(err, r)
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.AppsCount, err = sql.CountApps()
		log.Err(err, r)
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PackagesCount, err = sql.CountPackages()
		log.Err(err, r)
	}()

	// Get total prices
	wg.Add(1)
	go func() {

		defer wg.Done()

		t.Totals = map[string]string{}

		gorm, err := sql.GetMySQLClient()
		if err != nil {
			log.Err(err, r)
			return
		}

		var code = session.GetCountryCode(r)
		var rows []statsAppTypeTotalsRow

		gorm = gorm.Select([]string{"type", "round(sum(JSON_EXTRACT(prices, \"$." + string(code) + ".final\"))) as total"})
		gorm = gorm.Table("apps")
		gorm = gorm.Group("type")
		gorm = gorm.Where("type in (?)", []string{"game", "dlc"})
		gorm = gorm.Find(&rows)

		log.Err(gorm.Error, r)
		if gorm.Error == nil {

			for _, v := range rows {

				locale, err := helpers.GetLocaleFromCountry(code)
				log.Err(err, r)
				if err == nil {
					t.Totals[v.Type] = locale.Format(int(v.Total))
				}
			}
		}
	}()

	wg.Wait()

	err := returnTemplate(w, r, "stats", t)
	log.Err(err, r)
}

type statsTemplate struct {
	GlobalTemplate
	AppsCount     int
	PackagesCount int
	PlayersCount  int64
	Totals        map[string]string
}

type statsAppTypeTotalsRow struct {
	Type  string  `gorm:"column:type"`
	Total float64 `gorm:"column:total;type:float64"`
}

func statsClientPlayersHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Minute*30)

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.AddSelect("max(player_online)", "max_player_online")
	builder.SetFrom("GameDB", "alltime", "apps")
	builder.AddWhere("time", ">", "NOW() - 7d")
	builder.AddWhere("app_id", "=", "0")
	builder.AddGroupByTime("30m")
	builder.SetFillLinear()
	resp, err := helpers.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var hc helpers.HighChartsJson

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		hc = helpers.InfluxResponseToHighCharts(resp.Results[0].Series[0])
	}

	b, err := json.Marshal(hc)
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

func statsDatesHandler(w http.ResponseWriter, r *http.Request) {

	retu := setAllowedQueries(w, r, []string{})
	if retu {
		return
	}

	setCacheHeaders(w, time.Hour*6)

	gorm, err := sql.GetMySQLClient()
	if err != nil {

		log.Err(err, r)
		return
	}

	var dates []statsAppReleaseDate

	gorm = gorm.Select([]string{"count(*) as count", "release_date_unix as date"})
	gorm = gorm.Table("apps")
	gorm = gorm.Group("date")
	gorm = gorm.Order("date desc")
	gorm = gorm.Where("release_date_unix > ?", time.Now().AddDate(-1, 0, 0).Unix())
	gorm = gorm.Where("release_date_unix < ?", time.Now().AddDate(0, 0, 1).Unix())
	gorm = gorm.Find(&dates)

	log.Err(gorm.Error, r)

	var ret [][]int64
	for _, v := range dates {
		ret = append(ret, []int64{v.Date * 1000, int64(v.Count)})
	}

	bytes, err := json.Marshal(ret)
	log.Err(err, r)

	err = returnJSON(w, r, bytes)
	log.Err(err, r)
}

type statsAppReleaseDate struct {
	Date  int64
	Count int
}

func statsScoresHandler(w http.ResponseWriter, r *http.Request) {

	retu := setAllowedQueries(w, r, []string{})
	if retu {
		return
	}

	setCacheHeaders(w, time.Hour*24)

	gorm, err := sql.GetMySQLClient()
	if err != nil {

		log.Err(err, r)
		return
	}

	var scores []statsAppScore

	gorm = gorm.Select([]string{"FLOOR(reviews_score) AS score", "count(reviews_score) AS count"})
	gorm = gorm.Table("apps")
	gorm = gorm.Where("reviews_score > ?", 0)
	gorm = gorm.Group("FLOOR(reviews_score)")
	gorm = gorm.Find(&scores)

	log.Err(gorm.Error, r)

	ret := make([]int, 101) // 0-100
	for i := 0; i <= 100; i++ {
		ret[i] = 0
	}
	for _, v := range scores {
		ret[v.Score] = v.Count
	}

	bytes, err := json.Marshal(ret)
	log.Err(err, r)

	err = returnJSON(w, r, bytes)
	log.Err(err, r)
}

type statsAppScore struct {
	Score int
	Count int
}

func statsTypesHandler(w http.ResponseWriter, r *http.Request) {

	retu := setAllowedQueries(w, r, []string{})
	if retu {
		return
	}

	setCacheHeaders(w, time.Hour*24)

	gorm, err := sql.GetMySQLClient()
	if err != nil {

		log.Err(err, r)
		return
	}

	var types []statsAppType

	gorm = gorm.Select([]string{"type", "count(type) as count"})
	gorm = gorm.Table("apps")
	gorm = gorm.Group("type")
	gorm = gorm.Order("count desc")
	gorm = gorm.Find(&types)

	log.Err(gorm.Error, r)

	var ret [][]interface{}

	for _, v := range types {
		app := sql.App{}
		app.Type = v.Type
		ret = append(ret, []interface{}{app.GetType(), v.Count})
	}

	bytes, err := json.Marshal(ret)
	log.Err(err, r)

	err = returnJSON(w, r, bytes)
	log.Err(err, r)
}

type statsAppType struct {
	Type  string
	Count int
}
