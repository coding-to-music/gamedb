package web

import (
	"encoding/json"
	"net/http"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/logging"
)

func StatsHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := statsTemplate{}
	t.Fill(w, r, "Stats")
	t.Description = "Some interesting Steam Store stats"

	returnTemplate(w, r, "stats", t)
}

type statsTemplate struct {
	GlobalTemplate
}

func StatsScoresHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var scores []appScore
	gorm, err := db.GetMySQLClient()
	if err != nil {

		logging.Error(err)

	} else {

		gorm = gorm.Select([]string{"FLOOR(reviews_score) AS score", "count(reviews_score) AS count"})
		gorm = gorm.Table("apps")
		gorm = gorm.Where("reviews_score > ?", 0)
		gorm = gorm.Group("FLOOR(reviews_score)")
		//gorm = gorm.Order("reviews_score ASC")
		gorm = gorm.Find(&scores)

		logging.Error(gorm.Error)
	}

	ret := make([]int, 101) // 0-100
	for i := 0; i <= 100; i++ {
		ret[i] = 0
	}
	for _, v := range scores {
		ret[v.Score] = v.Count
	}

	bytes, err := json.Marshal(ret)
	logging.Error(err)

	w.Write(bytes)
}

type appScore struct {
	Score int
	Count int
}

func StatsTypesHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var types []appType
	gorm, err := db.GetMySQLClient()
	if err != nil {

		logging.Error(err)

	} else {

		gorm = gorm.Select([]string{"type", "count(type) as count"})
		gorm = gorm.Table("apps")
		gorm = gorm.Group("type")
		gorm = gorm.Order("count asc")
		gorm = gorm.Find(&types)

		logging.Error(gorm.Error)
	}

	var ret [][]interface{}

	for _, v := range types {
		app := db.App{}
		app.Type = v.Type
		ret = append(ret, []interface{}{app.GetType(), v.Count})
	}

	bytes, err := json.Marshal(ret)
	logging.Error(err)

	w.Write(bytes)
}

type appType struct {
	Type  string
	Count int
}

func StatsCountriesHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var types []appType
	gorm, err := db.GetMySQLClient()
	if err != nil {

		logging.Error(err)

	} else {

		gorm = gorm.Select([]string{"type", "count(type) as count"})
		gorm = gorm.Table("apps")
		gorm = gorm.Group("type")
		gorm = gorm.Order("count asc")
		gorm = gorm.Find(&types)

		logging.Error(gorm.Error)
	}

	var ret [][]interface{}

	for _, v := range types {
		app := db.App{}
		app.Type = v.Type
		ret = append(ret, []interface{}{app.GetType(), v.Count})
	}

	bytes, err := json.Marshal(ret)
	logging.Error(err)

	w.Write(bytes)
}
