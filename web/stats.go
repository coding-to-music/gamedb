package web

import (
	"encoding/json"
	"net/http"
	"sort"

	"cloud.google.com/go/datastore"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
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

	err := returnTemplate(w, r, "stats", t)
	log.Log(err)
}

type statsTemplate struct {
	GlobalTemplate
}

func statsScoresHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

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

	_, err = w.Write(bytes)
	log.Log(err)
}

type appScore struct {
	Score int
	Count int
}

func statsTypesHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

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

	_, err = w.Write(bytes)
	log.Log(err)
}

type appType struct {
	Type  string
	Count int
}

func statsCountriesHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

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

	_, err = w.Write(bytes)
	log.Log(err)
}
