package pages

import (
	"net/http"
	"strings"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
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
	t.fill(w, r, "Achievements", "")

	err := returnTemplate(w, r, "achievements", t)
	log.Err(err, r)
}

func achievementsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	query.limit(r)

	columns := map[string]string{
		"1": "achievements_count",
		"2": "achievements_average_completion",
	}

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err, r)
		return
	}

	gorm = gorm.Model(sql.App{})
	gorm = gorm.Select([]string{"id", "name", "icon", "achievements", "achievements_count", "achievements_average_completion", "prices"})
	gorm = gorm.Order(query.getOrderSQL(columns))
	gorm = gorm.Limit(100)
	gorm = gorm.Offset(query.getOffset())
	gorm = gorm.Where("achievements_count > 0")

	var apps []sql.App
	gorm = gorm.Find(&apps)
	log.Err(gorm.Error, r)

	count, err := sql.CountAppsWithAchievements()
	log.Err(err)

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = int64(count)
	response.RecordsFiltered = int64(count)
	response.Draw = query.Draw
	response.limit(r)

	var code = helpers.GetProductCC(r)

	for _, app := range apps {

		//noinspection GoPreferNilSlice
		var filteredIcons = []sql.AppAchievement{}

		icons, _ := app.GetAchievements()
		for _, v := range icons {
			if v.Active && strings.HasSuffix(v.Icon, ".jpg") {
				filteredIcons = append(filteredIcons, v)
				if len(filteredIcons) == 5 {
					break
				}
			}
		}

		response.AddRow([]interface{}{
			app.ID,                                // 0
			app.GetName(),                         // 1
			app.GetIcon(),                         // 2
			app.GetPath() + "#achievements",       // 3
			app.GetPrice(code).GetFinal(),         // 4
			app.AchievementsCount,                 // 5
			app.AchievementsAverageCompletion,     // 6
			filteredIcons,                         // 7
			helpers.InsertNewLines(app.GetName()), // 8
		})
	}

	response.output(w, r)
}
