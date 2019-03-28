package web

import (
	"net/http"
	"sync"
	"time"

	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
	"github.com/gamedb/website/sql"
)

// func homeRouter() http.Handler {
// 	r := chi.NewRouter()
// 	r.Get("/charts.json", homeChartsAjaxHandler)
// 	return r
// }

func homeHandler(w http.ResponseWriter, r *http.Request) {

	setCacheHeaders(w, time.Hour)

	t := homeTemplate{}
	t.fill(w, r, "Home", "Stats and information on the Steam Catalogue.")
	t.addAssetHighCharts()

	var wg sync.WaitGroup

	// wg.Add(1)
	// go func() {
	//
	// 	defer wg.Done()
	//
	// 	var err error
	// 	t.PlayersCount, err = mongo.CountPlayers()
	// 	log.Err(err, r)
	// }()
	//
	// wg.Add(1)
	// go func() {
	//
	// 	defer wg.Done()
	//
	// 	var err error
	// 	t.AppsCount, err = sql.CountApps()
	// 	log.Err(err, r)
	// }()
	//
	// wg.Add(1)
	// go func() {
	//
	// 	defer wg.Done()
	//
	// 	var err error
	// 	t.PackagesCount, err = sql.CountPackages()
	// 	log.Err(err, r)
	// }()

	// Popular
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PopularApps, err = sql.PopularApps()
		log.Err(err, r)
	}()

	// Trending
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.TrendingApps, err = sql.TrendingApps()
		log.Err(err, r)
	}()

	// New popular
	wg.Add(1)
	go func() {

		defer wg.Done()

		gorm, err := sql.GetMySQLClient()
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

		gorm, err := sql.GetMySQLClient()
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
	// AppsCount      int
	// PackagesCount  int
	// PlayersCount   int64
	PopularApps    []sql.App
	TrendingApps   []sql.App
	RatedNewApps   []sql.App
	PopularNewApps []sql.App
	Prices         map[int]string
}
