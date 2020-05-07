package pages

import (
	"net/http"
	"sync"

	"github.com/Jleagle/influxql"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers/i18n"
	"github.com/gamedb/gamedb/pkg/helpers/influx"
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
	t.addAssetJSON2HTML()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PlayersCount, err = mongo.CountDocuments(mongo.CollectionPlayers, nil, 0)
		if err != nil {
			log.Err(err, r)
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.AppsCount, err = mongo.CountDocuments(mongo.CollectionApps, nil, 0)
		if err != nil {
			log.Err(err, r)
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.BundlesCount, err = sql.CountBundles()
		if err != nil {
			log.Err(err, r)
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PackagesCount, err = mongo.CountDocuments(mongo.CollectionPackages, nil, 0)
		if err != nil {
			log.Err(err, r)
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		a := mongo.App{}
		t.SteamPlayersInGame, err = a.GetPlayersInGame()
		if err != nil {
			log.Err(err, r)
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		a := mongo.App{}
		t.SteamPlayersOnline, err = a.GetPlayersOnline()
		if err != nil {
			log.Err(err, r)
		}
	}()

	wg.Wait()

	returnTemplate(w, r, "stats", t)
}

type statsTemplate struct {
	GlobalTemplate
	AppsCount          int64
	BundlesCount       int
	PackagesCount      int64
	PlayersCount       int64
	SteamPlayersOnline int64
	SteamPlayersInGame int64
}

func playerLevelsHandler(w http.ResponseWriter, r *http.Request) {

	levels, err := mongo.GetPlayerLevelsRounded()
	if err != nil {
		log.Err(err, r)
		return
	}

	returnJSON(w, r, levels)
}

func statsAppTypesHandler(w http.ResponseWriter, r *http.Request) {

	var code = session.GetProductCC(r)
	var currency = i18n.GetProdCC(code).CurrencyCode

	types, err := mongo.GetAppsGroupedByType(code)
	if err != nil {
		log.Err(err, r)
		return
	}

	var resp statsAppTypes
	for _, v := range types {

		resp.Rows = append(resp.Rows, statsAppTypesRow{
			Type:           v.Type,
			Total:          v.Value,
			Count:          v.Count,
			TypeFormatted:  v.Format(),
			CountFormatted: humanize.Comma(v.Count),
			TotalFormatted: i18n.FormatPrice(currency, int(v.Value), true),
		})
	}

	// Get total
	var total int64
	for _, v := range resp.Rows {
		total += v.Total
	}
	resp.Total = i18n.FormatPrice(currency, int(total))

	//
	returnJSON(w, r, resp)
}

type statsAppTypes struct {
	Rows  []statsAppTypesRow `json:"rows"`
	Total string             `json:"total"`
}

type statsAppTypesRow struct {
	Type           string `json:"type"`
	Total          int64  `json:"total"`
	Count          int64  `json:"count"`
	TypeFormatted  string `json:"typef"`
	TotalFormatted string `json:"totalf"`
	CountFormatted string `json:"countf"`
}

func statsClientPlayersHandler(w http.ResponseWriter, r *http.Request) {

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.AddSelect("max(player_online)", "max_player_online")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW() - 7d")
	builder.AddWhere("app_id", "=", "0")
	builder.AddGroupByTime("30m")
	builder.SetFillNone()
	resp, err := influx.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	var hc influx.HighChartsJSON

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		hc = influx.InfluxResponseToHighCharts(resp.Results[0].Series[0], true)
	}

	returnJSON(w, r, hc)
}

func statsDatesHandler(w http.ResponseWriter, r *http.Request) {

	releaseDates, err := mongo.GetAppsGroupedByReleaseDate()
	if err != nil {
		log.Err(err, r)
		return
	}

	var ret [][]int64
	for _, v := range releaseDates {
		ret = append(ret, []int64{v.Date * 1000, v.Count})
	}

	returnJSON(w, r, ret)
}

func statsScoresHandler(w http.ResponseWriter, r *http.Request) {

	scores, err := mongo.GetAppsGroupedByReviewScore()
	if err != nil {
		log.Err(err, r)
		return
	}

	ret := make([]int64, 101) // 0-100
	for i := 0; i <= 100; i++ {
		ret[i] = 0
	}
	for _, v := range scores {
		ret[v.Score] = v.Count
	}

	returnJSON(w, r, ret)
}
