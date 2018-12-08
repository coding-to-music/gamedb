package web

import (
	"encoding/json"
	"net/http"
	"sort"
	"sync"

	"cloud.google.com/go/datastore"
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
	return r
}

func statsHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := statsTemplate{}
	t.Fill(w, r, "Stats", "Some interesting Steam Store stats.")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.RanksCount, err = db.CountRanks()
		log.Log(err)

	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.AppsCount, err = db.CountApps()
		log.Log(err)

	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PackagesCount, err = db.CountPackages()
		log.Log(err)

	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var rows []totalsRow
		gorm, err := db.GetMySQLClient(true)
		if err != nil {

			log.Log(err)
			return

		}

		code := session.GetCountryCode(r)

		gorm = gorm.Select([]string{"type", "round(sum(JSON_EXTRACT(prices, \"$." + string(code) + ".final\"))) as total"})
		gorm = gorm.Table("apps")
		gorm = gorm.Group("type")
		gorm = gorm.Order("total desc")
		gorm = gorm.Find(&rows)

		log.Log(gorm.Error)

		for _, v := range rows {

			locale, err := helpers.GetLocaleFromCountry(code)
			log.Log(err)

			final := locale.Format(v.Total)

			if v.Total > 0 && (v.Type == "game" || v.Type == "dlc") {
				app := db.App{}
				app.Type = v.Type
				t.Totals = append(t.Totals, totalsTemplate{
					"Total price of all " + app.GetType() + "s",
					final,
				})
			}
		}

	}()

	wg.Wait()

	err := returnTemplate(w, r, "stats", t)
	log.Log(err)
}

type statsTemplate struct {
	GlobalTemplate
	RanksCount    int
	AppsCount     int
	PackagesCount int
	Totals        []totalsTemplate
}

type totalsRow struct {
	Type  string `gorm:"column:type"`
	Total int    `gorm:"column:total;type:int"`
}

type totalsTemplate struct {
	Type  string
	Total string
}

func statsScoresHandler(w http.ResponseWriter, r *http.Request) {

	var scores []appScore
	gorm, err := db.GetMySQLClient()
	if err != nil {

		log.Log(err)

	} else {

		gorm = gorm.Select([]string{"FLOOR(reviews_score) AS score", "count(reviews_score) AS count"})
		gorm = gorm.Table("apps")
		gorm = gorm.Where("reviews_score > ?", 0)
		gorm = gorm.Group("FLOOR(reviews_score)")
		gorm = gorm.Find(&scores)

		log.Log(gorm.Error)
	}

	ret := make([]int, 101) // 0-100
	for i := 0; i <= 100; i++ {
		ret[i] = 0
	}
	for _, v := range scores {
		ret[v.Score] = v.Count
	}

	bytes, err := json.Marshal(ret)
	log.Log(err)

	err = returnJSON(w, r, bytes)
	log.Log(err)
}

type appScore struct {
	Score int
	Count int
}

func statsTypesHandler(w http.ResponseWriter, r *http.Request) {

	var types []appType
	gorm, err := db.GetMySQLClient()
	if err != nil {

		log.Log(err)

	} else {

		gorm = gorm.Select([]string{"type", "count(type) as count"})
		gorm = gorm.Table("apps")
		gorm = gorm.Group("type")
		gorm = gorm.Order("count desc")
		gorm = gorm.Find(&types)

		log.Log(gorm.Error)
	}

	var ret [][]interface{}

	for _, v := range types {
		app := db.App{}
		app.Type = v.Type
		ret = append(ret, []interface{}{app.GetType(), v.Count})
	}

	bytes, err := json.Marshal(ret)
	log.Log(err)

	err = returnJSON(w, r, bytes)
	log.Log(err)
}

type appType struct {
	Type  string
	Count int
}

func statsCountriesHandler(w http.ResponseWriter, r *http.Request) {

	var ranks []db.PlayerRank

	client, ctx, err := db.GetDSClient()
	log.Log(err)

	q := datastore.NewQuery(db.KindPlayerRank)

	_, err = client.GetAll(ctx, q, &ranks)
	if err != nil {
		log.Log(err)
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
	log.Log(err)

	err = returnJSON(w, r, bytes)
	log.Log(err)
}
