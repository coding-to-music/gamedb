package handlers

import (
	"net/http"
	"sync"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
)

func AchievementsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", achievementsHandler)
	r.Get("/achievements.json", achievementsAjaxHandler)
	return r
}

func achievementsHandler(w http.ResponseWriter, r *http.Request) {

	t := globalTemplate{}
	t.fill(w, r, "achievements", "Achievements", "Search all Steam achievements")

	returnTemplate(w, r, t)
}

func achievementsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	var query = datatable.NewDataTableQuery(r, true)
	var search = query.GetSearchString("search")

	var wg sync.WaitGroup

	var achievements []elasticsearch.Achievement
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		var sorters = query.GetOrderElastic(map[string]string{
			"1": "completed, app_owners desc",
		})

		achievements, filtered, err = elasticsearch.SearchAppAchievements(query.GetOffset(), search, sorters)
		if err != nil {
			log.ErrS(err)
		}
	}()

	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountDocuments(mongo.CollectionAppAchievements, nil, 60*60*24)
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Wait()

	//
	var response = datatable.NewDataTablesResponse(r, query, count, filtered, nil)
	for _, achievement := range achievements {

		response.AddRow([]interface{}{
			achievement.Name,          // 0
			achievement.GetIcon(),     // 1
			achievement.Description,   // 2
			achievement.GetCompleed(), // 3
			achievement.AppID,         // 4
			achievement.GetAppName(),  // 5
			achievement.Score,         // 6
			achievement.GetAppPath(),  // 7
			achievement.Hidden,        // 8
			achievement.NameMarked,    // 9
		})
	}

	returnJSON(w, r, response)
}
