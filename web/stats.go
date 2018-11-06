package web

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/logging"
)

func StatsHandler(w http.ResponseWriter, r *http.Request) {

	var wg sync.WaitGroup

	// Get app review scores
	var scores []appScore
	wg.Add(1)
	go func() {

		gorm, err := db.GetMySQLClient()
		if err != nil {

			logging.Error(err)

		} else {

			gorm = gorm.Select([]string{"round(reviews_score) AS score", "count(reviews_score) AS count"})
			gorm = gorm.Table("apps")
			gorm = gorm.Where("reviews_score > ?", 0)
			gorm = gorm.Group("round(reviews_score)")
			//gorm = gorm.Order("reviews_score ASC")
			gorm = gorm.Find(&scores)

			logging.Error(gorm.Error)
		}

		wg.Done()
	}()

	// Wait
	wg.Wait()

	// Template
	t := statsTemplate{}
	t.Fill(w, r, "Stats")
	t.setScoresJSON(scores)

	returnTemplate(w, r, "stats", t)
}

type statsTemplate struct {
	GlobalTemplate
	Scores string
}

func (s *statsTemplate) setScoresJSON(scores []appScore) {

	ret := make([]int, 101) // 0-100
	for i := 0; i <= 100; i++ {
		ret[i] = 0
	}
	for _, v := range scores {
		ret[v.Score] = v.Count
	}

	bytes, err := json.Marshal(ret)
	logging.Error(err)

	s.Scores = string(bytes)
}

type appScore struct {
	Score int
	Count int
}
