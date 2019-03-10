package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
	"github.com/go-chi/chi"
)

func upcomingRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", upcomingHandler)
	r.Get("/apps", upcomingAppsHandler)
	r.Get("/apps.json", upcomingAppsAjaxHandler)
	r.Get("/packages", upcomingPackagesHandler)
	r.Get("/packages.json", upcomingPackagesAjaxHandler)
	return r
}

func upcomingHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/upcoming/apps", 302)
}

func upcomingAppsHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	// Template
	t := upcomingTemplate{}
	t.Fill(w, r, "Upcoming Apps", "The apps you have to look forward to!")
	t.AjaxURL = "/upcoming/apps.json"

	t.Apps, err = countUpcomingApps()
	log.Err(err, r)

	t.Packages, err = countUpcomingPackages()
	log.Err(err, r)

	err = returnTemplate(w, r, "upcoming_apps", t)
	log.Err(err, r)
}

func upcomingPackagesHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	// Template
	t := upcomingTemplate{}
	t.Fill(w, r, "Upcoming Packages", "The packages you have to look forward to!")
	t.AjaxURL = "/upcoming/packages.json"

	t.Apps, err = countUpcomingApps()
	log.Err(err, r)

	t.Packages, err = countUpcomingPackages()
	log.Err(err, r)

	err = returnTemplate(w, r, "upcoming_packages", t)
	log.Err(err, r)
}

type upcomingTemplate struct {
	GlobalTemplate
	AjaxURL  string
	Apps     int
	Packages int
}

func upcomingAppsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.FillFromURL(r.URL.Query())
	log.Err(err, r)

	var count int
	var apps []db.App

	gorm, err := db.GetMySQLClient()
	if err != nil {

		log.Err(err, r)

	} else {

		gorm = gorm.Model(db.App{})
		gorm = gorm.Select([]string{"id", "name", "icon", "type", "prices", "release_date_unix"})
		gorm = gorm.Where("release_date_unix > ?", time.Now().AddDate(0, 0, -1).Unix())
		gorm = gorm.Order("release_date_unix asc, id asc")

		// Count before limitting
		gorm.Count(&count)
		log.Err(gorm.Error, r)

		gorm = gorm.Limit(100)
		gorm = gorm.Offset(query.Start)

		gorm = gorm.Find(&apps)
		log.Err(gorm.Error, r)
	}

	var code = session.GetCountryCode(r)

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(count)
	response.RecordsFiltered = strconv.Itoa(count)
	response.Draw = query.Draw

	for _, v := range apps {
		response.AddRow(v.OutputForJSONUpcoming(code))
	}

	response.output(w, r)
}

func upcomingPackagesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.FillFromURL(r.URL.Query())
	log.Err(err, r)

	var count int
	var packages []db.Package

	gorm, err := db.GetMySQLClient()
	if err != nil {

		log.Err(err, r)

	} else {

		gorm = gorm.Model(db.Package{})
		gorm = gorm.Select([]string{"id", "name", "apps_count", "prices", "release_date_unix"})
		gorm = gorm.Where("release_date_unix > ?", time.Now().AddDate(0, 0, -1).Unix())
		gorm = gorm.Order("release_date_unix asc, id asc")

		// Count before limitting
		gorm.Count(&count)
		log.Err(gorm.Error, r)

		gorm = gorm.Limit(100)
		gorm = gorm.Offset(query.Start)

		gorm = gorm.Find(&packages)
		log.Err(gorm.Error, r)
	}

	var code = session.GetCountryCode(r)

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(count)
	response.RecordsFiltered = strconv.Itoa(count)
	response.Draw = query.Draw

	for _, v := range packages {
		response.AddRow(v.OutputForJSONUpcoming(code))
	}

	response.output(w, r)
}

func countUpcomingPackages() (count int, err error) {

	var item = helpers.MemcacheUpcomingPackagesCount

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		var count int

		gorm, err := db.GetMySQLClient()
		if err != nil {
			return count, err
		}

		gorm = gorm.Model(db.Package{})
		gorm = gorm.Where("release_date_unix > ?", time.Now().AddDate(0, 0, -1).Unix())
		gorm = gorm.Count(&count)

		return count, gorm.Error
	})

	return count, err
}

func countUpcomingApps() (count int, err error) {

	var item = helpers.MemcacheUpcomingAppsCount

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		var count int

		gorm, err := db.GetMySQLClient()
		if err != nil {
			return count, err
		}

		gorm = gorm.Model(db.App{})
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
	log.Err(err)

	err = mc.Delete(helpers.MemcacheUpcomingPackagesCount.Key)
	log.Err(err)
}
