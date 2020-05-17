package pages

import (
	"net/http"
	"sync"

	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	elasticHelpers "github.com/gamedb/gamedb/pkg/elastic"
	"github.com/gamedb/gamedb/pkg/helpers"
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

	t := GlobalTemplate{}
	t.fill(w, r, "Achievements", "Search all Steam achievements")
	t.addAssetMark()

	returnTemplate(w, r, "achievements", t)
}

func achievementsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	var query = datatable.NewDataTableQuery(r, true)
	var search = query.GetSearchString("search")

	var wg sync.WaitGroup

	var achievements []elasticHelpers.Achievement
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		var sorters = query.GetOrderElastic(map[string]string{
			"1": "completed",
			"2": "_score",
		})

		achievements, filtered, err = elasticHelpers.SearchAchievements(100, query.GetOffset(), search, sorters)
		if err != nil {
			log.Err(err, r)
		}
	}()

	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountDocuments(mongo.CollectionAppAchievements, nil, 60*60*24)
		if err != nil {
			log.Err(err, r)
		}
	}()

	wg.Wait()

	//
	var response = datatable.NewDataTablesResponse(r, query, count, filtered)
	for _, achievement := range achievements {

		path := helpers.GetAppPath(achievement.AppID, achievement.AppName) + "#achievements"
		completed := helpers.FloatToString(achievement.Completed, 1)
		icon := helpers.GetAchievementIcon(achievement.AppID, achievement.Icon)
		appName := helpers.GetAppName(achievement.AppID, achievement.AppName)

		response.AddRow([]interface{}{
			achievement.Name,        // 0
			icon,                    // 1
			achievement.Description, // 2
			completed,               // 3
			achievement.AppID,       // 4
			appName,                 // 5
			achievement.Score,       // 6
			path,                    // 7
			achievement.Hidden,      // 8
		})
	}

	returnJSON(w, r, response)
}
