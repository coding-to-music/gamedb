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
	r.Get("/ajax", upcomingAjaxHandler)
	return r
}

func upcomingHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := upcomingTemplate{}
	t.Fill(w, r, "Upcoming", "All the apps you have to look forward to!")

	err := returnTemplate(w, r, "upcoming", t)
	log.Log(err)
}

type upcomingTemplate struct {
	GlobalTemplate
}

func upcomingAjaxHandler(w http.ResponseWriter, r *http.Request) {

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
		gorm = gorm.Order("release_date_unix asc, id asc")
		gorm = gorm.Where("release_date_unix > ?", time.Now().AddDate(0, 0, -1).Unix())

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
		response.AddRow(v.OutputForJSONComingSoon(code))
	}

	response.output(w)
}
