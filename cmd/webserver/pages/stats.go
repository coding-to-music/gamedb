package pages

import (
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
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
	r.Get("/app-types.json", statsAppTypesHandler)
	r.Get("/player-levels.json", playerLevelsHandler)
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
		t.BundlesCount, err = sql.CountBundles()
		log.Err(err, r)
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PackagesCount, err = sql.CountPackages()
		log.Err(err, r)
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		a := sql.App{}
		t.OnlinePlayersCount, err = a.GetOnlinePlayers()
		log.Err(err, r)
	}()

	wg.Wait()

	returnTemplate(w, r, "stats", t)
}

type statsTemplate struct {
	GlobalTemplate
	AppsCount          int
	BundlesCount       int
	PackagesCount      int
	PlayersCount       int64
	OnlinePlayersCount int64
}

func playerLevelsHandler(w http.ResponseWriter, r *http.Request) {

	levels, err := mongo.GetPlayerLevels()
	if err != nil {
		log.Err(err)
		return
	}

	returnJSON(w, r, levels)
}

func statsAppTypesHandler(w http.ResponseWriter, r *http.Request) {

	var ret statsAppTypes
	var code = helpers.GetProductCC(r)
	var item = memcache.MemcacheStatsAppTypes(code)

	err := memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &ret, func() (interface{}, error) {

		var code = helpers.GetProductCC(r)
		var rows []statsAppTypesRow

		gorm, err := sql.GetMySQLClient()
		if err != nil {
			return ret, err
		}

		gorm = gorm.Select([]string{"type", "count(type) as count", "round(sum(JSON_EXTRACT(prices, \"$." + string(code) + ".final\"))) as total"})
		gorm = gorm.Table("apps")
		gorm = gorm.Group("type")
		gorm = gorm.Find(&rows)

		if gorm.Error != nil {
			log.Err(gorm.Error, r)
			return ret, gorm.Error
		}

		for k := range rows {
			rows[k].TypeFormatted = helpers.GetAppType(rows[k].Type)
			rows[k].CountFormatted = humanize.Comma(rows[k].Count)
			rows[k].TotalFormatted = helpers.FormatPrice(helpers.GetProdCC(code).CurrencyCode, int(math.Round(rows[k].Total)), true)
		}

		// Get total
		var total float64
		for _, v := range rows {
			total += v.Total
		}

		ret = statsAppTypes{
			Rows:  rows,
			Total: helpers.FormatPrice(helpers.GetProdCC(code).CurrencyCode, int(total)),
		}

		return ret, nil
	})
	if err != nil {
		log.Err(err)
	}

	returnJSON(w, r, ret)
}

type statsAppTypes struct {
	Rows  []statsAppTypesRow `json:"rows"`
	Total string             `json:"total"`
}

type statsAppTypesRow struct {
	Type           string  `gorm:"column:type" json:"type"`
	Total          float64 `gorm:"column:total;type:float64" json:"total"`
	Count          int64   `gorm:"column:count;type:int" json:"count"`
	TypeFormatted  string  `json:"typef"`
	TotalFormatted string  `json:"totalf"`
	CountFormatted string  `json:"countf"`
}

func statsClientPlayersHandler(w http.ResponseWriter, r *http.Request) {

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.AddSelect("max(player_online)", "max_player_online")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW() - 7d")
	builder.AddWhere("app_id", "=", "0")
	builder.AddGroupByTime("30m")
	builder.SetFillLinear()
	resp, err := influx.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var hc influx.HighChartsJSON

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		hc = influx.InfluxResponseToHighCharts(resp.Results[0].Series[0])
	}

	returnJSON(w, r, hc)
}

func statsDatesHandler(w http.ResponseWriter, r *http.Request) {

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

	returnJSON(w, r, ret)
}

type statsAppReleaseDate struct {
	Date  int64
	Count int
}

func statsScoresHandler(w http.ResponseWriter, r *http.Request) {

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

	returnJSON(w, r, ret)
}

type statsAppScore struct {
	Score int
	Count int
}
