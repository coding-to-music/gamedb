package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Jleagle/memcache-go/memcache"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
	"github.com/gamedb/website/sql"
	"github.com/go-chi/chi"
)

func upcomingRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", upcomingHandler)
	r.Get("/upcoming.json", upcomingAjaxHandler)
	return r
}

func upcomingHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*12)

	var err error

	// Template
	t := upcomingTemplate{}
	t.fill(w, r, "Upcoming", "The apps you have to look forward to!")

	t.Apps, err = countUpcomingApps()
	log.Err(err, r)

	err = returnTemplate(w, r, "upcoming", t)
	log.Err(err, r)
}

type upcomingTemplate struct {
	GlobalTemplate
	Apps int
}

func upcomingAjaxHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{"draw"})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*1)

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err, r)
		return
	}

	gorm = gorm.Model(sql.App{})
	gorm = gorm.Select([]string{"id", "name", "icon", "type", "prices", "release_date_unix"})
	gorm = gorm.Where("release_date_unix >= ?", time.Now().AddDate(0, 0, -1).Unix())
	gorm = gorm.Order("release_date_unix ASC, name ASC")

	if config.Config.IsLocal() {
		gorm = gorm.Limit(100)
	} else {
		gorm = gorm.Limit(10000)
	}

	gorm = gorm.Offset(query.Start)

	var apps []sql.App
	gorm = gorm.Find(&apps)
	log.Err(gorm.Error, r)

	var code = session.GetCountryCode(r)

	count, err := countUpcomingApps()
	log.Err(err)

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(count)
	response.RecordsFiltered = strconv.Itoa(count)
	response.Draw = query.Draw

	for _, app := range apps {
		response.AddRow([]interface{}{
			app.ID,
			app.GetName(),
			app.GetIcon(),
			app.GetPath(),
			app.GetType(),
			sql.GetPriceFormatted(app, code).Final,
			app.GetDaysToRelease() + " (" + app.GetReleaseDateNice() + ")",
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
		gorm = gorm.Where("release_date_unix > ?", time.Now().AddDate(0, 0, -1).Unix())
		gorm = gorm.Count(&count)

		return count, gorm.Error
	})

	return count, err
}

func ClearUpcomingCache() {

	var mc = helpers.GetMemcache()
	var err error

	err = mc.Delete(helpers.MemcacheUpcomingAppsCount.Key)
	err = helpers.IgnoreErrors(err, memcache.ErrCacheMiss)
	log.Err(err)

	err = mc.Delete(helpers.MemcacheUpcomingPackagesCount.Key)
	err = helpers.IgnoreErrors(err, memcache.ErrCacheMiss)
	log.Err(err)
}
