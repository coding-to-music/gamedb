package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/logging"
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
	t.Fill(w, r, "Upcoming Apps")

	returnTemplate(w, r, "upcoming", t)
}

type upcomingTemplate struct {
	GlobalTemplate
}

func upcomingAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	query.FillFromURL(r.URL.Query())

	var count int
	var apps []db.App

	gorm, err := db.GetMySQLClient()
	if err != nil {

		logging.Error(err)

	} else {

		gorm = gorm.Model(db.App{})
		gorm = gorm.Select([]string{"id", "name", "icon", "type", "prices", "release_date_unix"})
		gorm = gorm.Order("release_date_unix desc")
		gorm = gorm.Where("release_date_unix > ?", time.Now().Unix())

		// Count before limitting
		gorm.Count(&count)
		logging.Error(gorm.Error)

		gorm = gorm.Limit(100)
		gorm = gorm.Offset(query.Start)

		gorm = gorm.Find(&apps)
		logging.Error(gorm.Error)
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
