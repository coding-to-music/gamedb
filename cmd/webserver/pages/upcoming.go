package pages

import (
	"net/http"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
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

	db, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err, r)
		return
	}

	search := query.getSearchString("search")
	filtered := 0

	db = db.Model(sql.App{})
	db = db.Select([]string{"id", "name", "icon", "type", "prices", "release_date_unix", "group_id", "group_followers"})
	db = db.Where("release_date_unix >= ?", time.Now().AddDate(0, 0, -1).Unix())
	if search != "" {
		db = db.Where("name LIKE ?", "%"+search+"%")
		db = db.Count(&filtered)
		log.Err(db.Error, r)
	}

	sortCols := map[string]string{
		"1": "group_followers $dir, name ASC",
		"4": "release_date_unix $dir, group_followers DESC, name ASC",
	}
	db = query.setOrderOffsetGorm(db, sortCols, "4")

	db = db.Limit(100)

	var apps []sql.App
	db = db.Find(&apps)
	log.Err(db.Error, r)

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
			app.ID,                          // 0
			app.GetName(),                   // 1
			app.GetIcon(),                   // 2
			app.GetPath(),                   // 3
			app.GetType(),                   // 4
			app.GetPrice(code).GetFinal(),   // 5
			app.GetReleaseDateNice(),        // 6
			app.GetFollowers(),              // 7
			helpers.GetAppStoreLink(app.ID), // 8
			app.ReleaseDateUnix,             // 9
			time.Unix(app.ReleaseDateUnix, 0).Format(helpers.DateYear), // 10
		})
	}

	response.output(w, r)
}

func countUpcomingApps() (count int, err error) {

	var item = memcache.MemcacheUpcomingAppsCount

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

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
