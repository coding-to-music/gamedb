package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
	"github.com/go-chi/chi"
)

func upcomingRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", upcomingHandler)
	r.Get("/apps", upcomingAppsHandler)
	r.Get("/apps/ajax", upcomingAppsAjaxHandler)
	r.Get("/packages", upcomingPackagesHandler)
	r.Get("/packages/ajax", upcomingPackagesAjaxHandler)
	return r
}

func upcomingHandler(w http.ResponseWriter, r *http.Request) {

	http.Redirect(w, r, "/upcoming/apps", 302)
	return
}

func upcomingAppsHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	// Template
	t := upcomingTemplate{}
	t.Fill(w, r, "Upcoming Apps", "The apps you have to look forward to!")
	t.AjaxURL = "/upcoming/apps/ajax"

	t.Apps, err = db.CountUpcomingApps()
	log.Log(err)

	t.Packages, err = db.CountUpcomingPackages()
	log.Log(err)

	err = returnTemplate(w, r, "upcoming_apps", t)
	log.Log(err)
}

func upcomingPackagesHandler(w http.ResponseWriter, r *http.Request) {

	var err error

	// Template
	t := upcomingTemplate{}
	t.Fill(w, r, "Upcoming Packages", "The packages you have to look forward to!")
	t.AjaxURL = "/upcoming/packages/ajax"

	t.Apps, err = db.CountUpcomingApps()
	log.Log(err)

	t.Packages, err = db.CountUpcomingPackages()
	log.Log(err)

	err = returnTemplate(w, r, "upcoming_packages", t)
	log.Log(err)
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
	log.Log(err)

	var count int
	var apps []db.App

	gorm, err := db.GetMySQLClient()
	if err != nil {

		log.Log(err)

	} else {

		gorm = gorm.Model(db.App{})
		gorm = gorm.Select([]string{"id", "name", "icon", "type", "prices", "release_date_unix"})
		gorm = gorm.Where("release_date_unix > ?", time.Now().AddDate(0, 0, -1).Unix())
		gorm = gorm.Order("release_date_unix asc, id asc")

		// Count before limitting
		gorm.Count(&count)
		log.Log(gorm.Error)

		gorm = gorm.Limit(100)
		gorm = gorm.Offset(query.Start)

		gorm = gorm.Find(&apps)
		log.Log(gorm.Error)
	}

	var code = session.GetCountryCode(r)

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(count)
	response.RecordsFiltered = strconv.Itoa(count)
	response.Draw = query.Draw

	for _, v := range apps {
		response.AddRow(v.OutputForJSONUpcoming(code))
	}

	response.output(w)
}

func upcomingPackagesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.FillFromURL(r.URL.Query())
	log.Log(err)

	var count int
	var packages []db.Package

	gorm, err := db.GetMySQLClient()
	if err != nil {

		log.Log(err)

	} else {

		gorm = gorm.Model(db.Package{})
		gorm = gorm.Select([]string{"id", "name", "apps_count", "prices", "release_date_unix"})
		gorm = gorm.Where("release_date_unix > ?", time.Now().AddDate(0, 0, -1).Unix())
		gorm = gorm.Order("release_date_unix asc, id asc")

		// Count before limitting
		gorm.Count(&count)
		log.Log(gorm.Error)

		gorm = gorm.Limit(100)
		gorm = gorm.Offset(query.Start)

		gorm = gorm.Find(&packages)
		log.Log(gorm.Error)
	}

	var code = session.GetCountryCode(r)

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(count)
	response.RecordsFiltered = strconv.Itoa(count)
	response.Draw = query.Draw

	for _, v := range packages {
		response.AddRow(v.OutputForJSONUpcoming(code))
	}

	response.output(w)
}
