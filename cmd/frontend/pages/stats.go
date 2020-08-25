package pages

import (
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/cmd/frontend/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func StatsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", statsHandler)
	r.Get("/gamedb", gameDBStatsHandler)
	r.Get("/client-players.json", statsClientPlayersHandler)
	r.Get("/client-players2.json", statsClientPlayers2Handler)
	r.Get("/release-dates.json", statsDatesHandler)
	r.Get("/app-scores.json", statsScoresHandler)
	r.Get("/app-types.json", statsAppTypesHandler)
	r.Get("/player-levels.json", playerLevelsHandler)
	r.Get("/player-countries.json", playerCountriesHandler)
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
		t.AppsCount, err = mongo.CountDocuments(mongo.CollectionApps, nil, 0)
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.BundlesCount, err = mysql.CountBundles()
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PackagesCount, err = mongo.CountDocuments(mongo.CollectionPackages, nil, 0)
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.AchievementsCount, err = mongo.CountDocuments(mongo.CollectionAppAchievements, nil, 0)
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.ArticlesCount, err = mongo.CountDocuments(mongo.CollectionAppArticles, nil, 0)
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		a := mongo.App{}
		t.SteamPlayersInGame, err = a.GetPlayersInGame()
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		a := mongo.App{}
		t.SteamPlayersOnline, err = a.GetPlayersOnline()
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Wait()

	returnTemplate(w, r, "stats_steam", t)
}

type statsTemplate struct {
	globalTemplate
	AppsCount          int64
	BundlesCount       int
	PackagesCount      int64
	AchievementsCount  int64
	ArticlesCount      int64
	SteamPlayersOnline int64
	SteamPlayersInGame int64
}

func (t statsTemplate) includes() []string {
	return []string{"includes/stats_header.gohtml"}
}

func gameDBStatsHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := gamedbStatsTemplate{}
	t.fill(w, r, "Stats", "Some interesting Steam Store stats.")
	t.addAssetHighCharts()
	t.addAssetHighChartsDrilldown()
	t.addAssetJSON2HTML()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PlayersCount, err = mongo.CountDocuments(mongo.CollectionPlayers, nil, 0)
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PlayerAppsCount, err = mongo.CountDocuments(mongo.CollectionPlayerApps, nil, 0)
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PlayerFriendsCount, err = mongo.CountDocuments(mongo.CollectionPlayerFriends, nil, 0)
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PlayerAchievementsCount, err = mongo.CountDocuments(mongo.CollectionPlayerAchievements, nil, 0)
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PlayerBadgesCount, err = mongo.CountDocuments(mongo.CollectionPlayerBadges, nil, 0)
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PlayerGroupsCount, err = mongo.CountDocuments(mongo.CollectionPlayerGroups, nil, 0)
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Wait()

	returnTemplate(w, r, "stats_gamedb", t)
}

type gamedbStatsTemplate struct {
	globalTemplate
	PlayerAppsCount         int64
	PlayerFriendsCount      int64
	PlayerAchievementsCount int64
	PlayerBadgesCount       int64
	PlayerGroupsCount       int64
	PlayersCount            int64
}

func (t gamedbStatsTemplate) includes() []string {
	return []string{"includes/stats_header.gohtml"}
}

func playerLevelsHandler(w http.ResponseWriter, r *http.Request) {

	levels, err := mongo.GetPlayerLevelsRounded()
	if err != nil {
		log.ErrS(err)
		return
	}

	returnJSON(w, r, levels)
}

func playerCountriesHandler(w http.ResponseWriter, r *http.Request) {

	aggs, err := elasticsearch.AggregatePlayerCountries()
	if err != nil {
		log.ErrS(err)
		return
	}

	var series []playerCountrySeries
	var drilldown []map[string]interface{}

	for cc := range i18n.States {

		if aggs[cc] == 0 {
			continue
		}

		countryName := i18n.CountryCodeToName(cc)

		series = append(series, playerCountrySeries{
			Name:  countryName,
			ID:    cc,
			Value: aggs[cc],
		})

		var data [][]interface{}
		for sc, name := range i18n.States[cc] {
			if aggs[cc+"-"+sc] > 0 {
				data = append(data, []interface{}{name, aggs[cc+"-"+sc]})
			}
		}

		sort.Slice(data, func(i, j int) bool {
			return data[i][1].(int64) > data[j][1].(int64)
		})

		drilldown = append(drilldown, map[string]interface{}{
			"name": countryName,
			"id":   cc,
			"data": data,
		})
	}

	sort.Slice(series, func(i, j int) bool {
		return series[i].Value > series[j].Value
	})

	if len(series) > 40 {
		series = series[0:40]
	}

	returnJSON(w, r, map[string]interface{}{
		"series":    series,
		"drilldown": drilldown,
	})
}

type playerCountrySeries struct {
	Name  string `json:"name"`
	ID    string `json:"drilldown"`
	Value int64  `json:"y"`
}

func statsAppTypesHandler(w http.ResponseWriter, r *http.Request) {

	var code = session.GetProductCC(r)
	var currency = i18n.GetProdCC(code).CurrencyCode

	types, err := mongo.GetAppsGroupedByType(code)
	if err != nil {
		log.ErrS(err)
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
	builder.AddGroupByTime("10m")
	builder.SetFillNone()

	resp, err := influx.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err.Error(), zap.String("query", builder.String()))
		return
	}

	var hc influx.HighChartsJSON

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		hc = influx.InfluxResponseToHighCharts(resp.Results[0].Series[0], false)
	}

	returnJSON(w, r, hc)
}

func statsClientPlayers2Handler(w http.ResponseWriter, r *http.Request) {

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.AddSelect("max(player_online)", "max_player_online")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
	// builder.AddWhere("time", ">", "NOW()-1825d")
	builder.AddWhere("time", ">", "2019-05-04") // Bad data befoe this
	builder.AddWhere("app_id", "=", "0")
	builder.AddGroupByTime("1d")
	builder.SetFillNumber(0)

	resp, err := influx.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err.Error(), zap.String("query", builder.String()))
		return
	}

	var hc influx.HighChartsJSON

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {

		hc = influx.InfluxResponseToHighCharts(resp.Results[0].Series[0], false)
	}

	returnJSON(w, r, hc)
}

func statsDatesHandler(w http.ResponseWriter, r *http.Request) {

	releaseDates, err := mongo.GetAppsGroupedByReleaseDate()
	if err != nil {
		log.ErrS(err)
		return
	}

	var ret [][]int64
	for _, v := range releaseDates {
		ts, _ := time.Parse("2006-01-02", v.Date)
		ret = append(ret, []int64{ts.Unix() * 1000, v.Count})
	}

	returnJSON(w, r, ret)
}

func statsScoresHandler(w http.ResponseWriter, r *http.Request) {

	scores, err := mongo.GetAppsGroupedByReviewScore()
	if err != nil {
		log.ErrS(err)
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
