package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/influxql"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
	"github.com/go-chi/chi"
)

func statsRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", statsHandler)
	r.Get("/app-scores", statsScoresHandler)
	r.Get("/app-types", statsTypesHandler)
	r.Get("/ranked-countries", statsCountriesHandler)
	r.Get("/release-dates", statsDatesHandler)
	r.Get("/client-players", statsClientPlayersHandler)
	return r
}

func statsHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := statsTemplate{}
	t.fill(w, r, "Stats", "Some interesting Steam Store stats.")
	t.addAssetHighCharts()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.RanksCount, err = db.CountRanks()
		log.Err(err, r)

	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.AppsCount, err = db.CountApps()
		log.Err(err, r)

	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PackagesCount, err = db.CountPackages()
		log.Err(err, r)

	}()

	// Get total prices
	wg.Add(1)
	go func() {

		defer wg.Done()

		gorm, err := db.GetMySQLClient()
		if err != nil {

			log.Err(err, r)
			return

		}

		code := session.GetCountryCode(r)

		gorm = gorm.Select([]string{"type", "round(sum(JSON_EXTRACT(prices, \"$." + string(code) + ".final\"))) as total"})
		gorm = gorm.Table("apps")
		gorm = gorm.Where("type in (?)", []string{"game", "dlc"})
		gorm = gorm.Group("type")
		gorm = gorm.Order("total desc")

		var rows []statsAppTypeTotalsRow
		gorm = gorm.Find(&rows)
		log.Err(gorm.Error, r)

		fmt.Println(rows)

		t.Totals = map[string]string{}
		for _, v := range rows {

			locale, err := helpers.GetLocaleFromCountry(code)
			log.Err(err, r)
			if err == nil {
				t.Totals[v.Type] = locale.Format(int(v.Total))
			}
		}
	}()

	wg.Wait()

	err := returnTemplate(w, r, "stats", t)
	log.Err(err, r)
}

type statsTemplate struct {
	GlobalTemplate
	RanksCount    int
	AppsCount     int
	PackagesCount int
	Totals        map[string]string
}

type statsAppTypeTotalsRow struct {
	Type  string  `gorm:"column:type"`
	Total float64 `gorm:"column:total;type:float64"`
}

func statsScoresHandler(w http.ResponseWriter, r *http.Request) {

	gorm, err := db.GetMySQLClient()
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

	gorm, err := db.GetMySQLClient()
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
		app := db.App{}
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

func statsCountriesHandler(w http.ResponseWriter, r *http.Request) {

	var ranks []db.PlayerRank

	client, ctx, err := db.GetDSClient()
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Error: err, Code: 500, Message: "Something went wrong"})
		return
	}

	q := datastore.NewQuery(db.KindPlayerRank)

	if config.Config.IsLocal() {
		q = q.Limit(1000)
	}

	_, err = client.GetAll(ctx, q, &ranks)
	if err != nil {
		log.Err(err, r)
	}

	// Tally up
	tally := map[string]int{}
	for _, v := range ranks {
		if _, ok := tally[v.CountryCode]; ok {
			tally[v.CountryCode]++
		} else {
			tally[v.CountryCode] = 1
		}
	}

	// Filter
	for k, v := range tally {
		if v < 10 {
			delete(tally, k)
		}
	}

	var ret [][]interface{}

	for k, v := range tally {
		if k == "" {
			k = "??"
		}
		ret = append(ret, []interface{}{k, v})
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i][1].(int) > ret[j][1].(int)
	})

	bytes, err := json.Marshal(ret)
	log.Err(err, r)

	err = returnJSON(w, r, bytes)
	log.Err(err, r)
}

func statsDatesHandler(w http.ResponseWriter, r *http.Request) {

	gorm, err := db.GetMySQLClient()
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

func statsClientPlayersHandler(w http.ResponseWriter, r *http.Request) {

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom("GameDB", "alltime", "apps")
	builder.AddWhere("time", ">", "NOW() - 30d")
	builder.AddWhere("app_id", "=", 0)
	builder.AddGroupByTime("30m")
	builder.SetFillNone()

	resp, err := db.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var hc db.HighChartsJson

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		hc = db.InfluxResponseToHighCharts(resp.Results[0].Series[0])
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
