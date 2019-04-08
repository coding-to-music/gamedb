package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
	"github.com/gamedb/website/sql"
	"github.com/go-chi/chi"
)

func newReleasesRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", newReleasesHandler)
	r.Get("/new-releases.json", newReleasesAjaxHandler)
	return r
}

func newReleasesHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	t := newReleasesTemplate{}
	t.fill(w, r, "New Releases", "")
	t.addAssetHighCharts()

	t.Apps, err = countUpcomingApps()
	log.Err(err, r)

	err = returnTemplate(w, r, "new_releases", t)
	log.Err(err, r)
}

type newReleasesTemplate struct {
	GlobalTemplate
	Apps int
}

func newReleasesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	var count int
	var apps []sql.App

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err, r)
		return
	}

	columns := map[string]string{
		"0": "name",
		"1": "price",
		"2": "reviews_score",
		"3": "player_peak_week",
		"4": "release_date_unix",
		"5": "player_trend",
	}

	var code = session.GetCountryCode(r)

	gorm = gorm.Model(sql.App{})
	gorm = gorm.Select([]string{"id", "name", "icon", "type", "prices", "release_date_unix", "player_peak_week", "reviews_score"})
	gorm = gorm.Where("release_date_unix < ?", time.Now().Unix())
	gorm = gorm.Where("release_date_unix > ?", time.Now().Add(time.Hour * 24 * 7 * -1).Unix())
	gorm = gorm.Order(query.getOrderSQL(columns, code))

	// Count before limitting
	gorm.Count(&count)
	log.Err(gorm.Error, r)

	gorm = gorm.Limit(50)
	gorm = gorm.Offset(query.getOffset())

	gorm = gorm.Find(&apps)
	log.Err(gorm.Error, r)

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(count)
	response.RecordsFiltered = strconv.Itoa(count)
	response.Draw = query.Draw

	for _, app := range apps {
		response.AddRow([]interface{}{
			app.ID,                                 // 0
			app.GetName(),                          // 1
			app.GetIcon(),                          // 2
			app.GetPath(),                          // 3
			app.GetType(),                          // 4
			sql.GetPriceFormatted(app, code).Final, // 5
			app.GetReleaseDateNice(),               // 6
			helpers.RoundFloatTo2DP(app.ReviewsScore), // 7
			app.PlayerPeakWeek,                        // 8
		})
	}

	response.output(w, r)
}
