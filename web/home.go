package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
	"github.com/go-chi/chi"
)

func homeRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/charts.json", homeChartsAjaxHandler)
	return r
}

func homeHandler(w http.ResponseWriter, r *http.Request) {

	t := homeTemplate{}
	t.fill(w, r, "Home", "Stats and information on the Steam Catalogue.")
	t.addAssetHighCharts()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PlayersCount, err = db.CountPlayers()
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

	// Popular
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PopularApps, err = db.PopularApps()
		log.Err(err, r)
	}()

	// Trending
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.TrendingApps, err = db.TrendingApps()
		log.Err(err, r)
	}()

	// New popular
	wg.Add(1)
	go func() {

		defer wg.Done()

		gorm, err := db.GetMySQLClient()
		if err != nil {
			log.Err(err)
			return
		}

		gorm = gorm.Select([]string{"id", "name", "icon", "player_peak_week", "prices"})
		gorm = gorm.Where("type = ?", "game")
		gorm = gorm.Where("release_date_unix > ?", time.Now().Add(time.Hour * 24 * 7 * -1).Unix())
		gorm = gorm.Order("player_peak_week desc")
		gorm = gorm.Limit(15)
		gorm = gorm.Find(&t.PopularNewApps)

		log.Err(err, r)
	}()

	// New rated
	wg.Add(1)
	go func() {

		defer wg.Done()

		gorm, err := db.GetMySQLClient()
		if err != nil {
			log.Err(err)
			return
		}

		gorm = gorm.Select([]string{"id", "name", "icon", "reviews_score", "prices"})
		gorm = gorm.Where("type = ?", "game")
		gorm = gorm.Where("release_date_unix > ?", time.Now().Add(time.Hour * 24 * 7 * -1).Unix())
		gorm = gorm.Order("reviews_score desc")
		gorm = gorm.Limit(15)
		gorm = gorm.Find(&t.RatedNewApps)

		log.Err(err, r)
	}()

	wg.Wait()

	// Get prices
	var prices = map[int]string{}
	var code = session.GetCountryCode(r)

	for _, v := range t.PopularNewApps {
		p, err := v.GetPrice(code)
		log.Err(err)
		if err == nil {
			prices[v.ID] = p.GetFinal()
		}
	}

	for _, v := range t.RatedNewApps {
		p, err := v.GetPrice(code)
		log.Err(err)
		if err == nil {
			prices[v.ID] = p.GetFinal()
		}
	}

	t.Prices = prices

	//
	err := returnTemplate(w, r, "home", t)
	log.Err(err, r)
}

type homeTemplate struct {
	GlobalTemplate
	AppsCount      int
	PackagesCount  int
	PlayersCount   int
	PopularApps    []db.App
	TrendingApps   []db.App
	RatedNewApps   []db.App
	PopularNewApps []db.App
	Prices         map[int]string
}

func homeChartsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	apps, err := db.TrendingApps()
	if err != nil {
		log.Err(err, r)
		return
	}

	if len(apps) == 0 {
		return
	}

	var or []string
	for _, v := range apps {
		or = append(or, `"app_id" = '`+strconv.Itoa(v.ID)+`'`)
	}

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom("GameDB", "alltime", "apps")
	builder.AddWhere("time", ">", "NOW()-7d")
	builder.AddWhereRaw("(" + strings.Join(or, " OR ") + ")")
	builder.AddGroupByTime("30m")
	builder.AddGroupBy("app_id")
	builder.SetFillNone()

	resp, err := db.InfluxQuery(builder.String())
	if err != nil {
		log.Err(err, r, builder.String())
		return
	}

	ret := map[string]db.HighChartsJson{}
	if len(resp.Results) > 0 {
		for _, v := range resp.Results[0].Series {
			ret[v.Tags["app_id"]] = db.InfluxResponseToHighCharts(v)
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
