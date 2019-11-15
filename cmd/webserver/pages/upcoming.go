package pages

import (
	"net/http"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
)

func UpcomingRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", upcomingHandler)
	r.Get("/upcoming.json", upcomingAjaxHandler)
	return r
}

func upcomingHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	// Template
	t := upcomingTemplate{}
	t.fill(w, r, "Upcoming", "The apps you have to look forward to!")

	t.Apps, err = countUpcomingApps()
	log.Err(err, r)

	returnTemplate(w, r, "upcoming", t)
}

type upcomingTemplate struct {
	GlobalTemplate
	Apps int
}

func upcomingAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err, r)
		return
	}

	search := query.getSearchString("search")
	filtered := 0

	gorm = gorm.Model(sql.App{})
	gorm = gorm.Select([]string{"id", "name", "icon", "type", "prices", "release_date_unix", "group_followers"})
	gorm = gorm.Where("release_date_unix >= ?", time.Now().AddDate(0, 0, -1).Unix())
	if search != "" {
		gorm = gorm.Where("name LIKE ?", "%"+search+"%")
		gorm = gorm.Count(&filtered)
		log.Err(gorm.Error, r)
	}
	gorm = gorm.Order("release_date_unix ASC, group_followers DESC, name ASC")
	gorm = gorm.Limit(100)
	gorm = gorm.Offset(query.getOffset())

	var apps []sql.App
	gorm = gorm.Find(&apps)
	log.Err(gorm.Error, r)

	var code = helpers.GetProductCC(r)

	count, err := countUpcomingApps()
	log.Err(err)

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = int64(count)
	response.RecordsFiltered = int64(count)
	if search != "" {
		response.RecordsFiltered = int64(filtered)
	}
	response.Draw = query.Draw

	for _, app := range apps {

		response.AddRow([]interface{}{
			app.ID,                        // 0
			app.GetName(),                 // 1
			app.GetIcon(),                 // 2
			app.GetPath(),                 // 3
			app.GetType(),                 // 4
			app.GetPrice(code).GetFinal(), // 5
			app.GetDaysToRelease() + " (" + app.GetReleaseDateNice() + ")", // 6
			app.GroupFollowers,              // 7
			helpers.GetAppStoreLink(app.ID), // 8
		})
	}

	response.output(w, r)
}

func countUpcomingApps() (count int, err error) {

	var item = helpers.MemcacheUpcomingAppsCount

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		var count int

		gorm, err := sql.GetMySQLClient()
		if err != nil {
			return count, err
		}

		gorm = gorm.Model(sql.App{})
		gorm = gorm.Where("release_date_unix >= ?", time.Now().AddDate(0, 0, -1).Unix())
		gorm = gorm.Count(&count)

		return count, gorm.Error
	})

	return count, err
}
