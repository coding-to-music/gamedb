package web

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/logging"
	"github.com/gamedb/website/session"
	"github.com/go-chi/chi"
)

func packagesRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", PackagesHandler)
	r.Get("/ajax", PackagesAjaxHandler)
	r.Get("/{id}", PackageHandler)
	r.Get("/{id}/{slug}", PackageHandler)
	return r
}

func PackagesHandler(w http.ResponseWriter, r *http.Request) {

	total, err := db.CountPackages()
	logging.Error(err)

	// Template
	t := packagesTemplate{}
	t.Fill(w, r, "Packages")
	t.Description = "The last " + humanize.Comma(int64(total)) + " packages to be updated."

	err = returnTemplate(w, r, "packages", t)
	logging.Error(err)
}

type packagesTemplate struct {
	GlobalTemplate
}

func PackagesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	query := DataTablesQuery{}
	err := query.FillFromURL(r.URL.Query())
	logging.Error(err)

	//
	var wg sync.WaitGroup

	// Get apps
	var packages []db.Package

	wg.Add(1)
	go func(r *http.Request) {

		gorm, err := db.GetMySQLClient()
		if err != nil {

			logging.Error(err)

		} else {

			gorm = gorm.Model(&db.Package{})
			gorm = gorm.Select([]string{"id", "name", "billing_type", "license_type", "status", "apps_count", "change_number_date"})

			gorm = query.SetOrderOffsetGorm(gorm, session.GetCountryCode(r), map[string]string{
				"0": "name",
				"4": "apps_count",
				"5": "change_number_date",
			})

			gorm = gorm.Limit(100)
			gorm = gorm.Find(&packages)

			logging.Error(gorm.Error)
		}

		wg.Done()
	}(r)

	// Get total
	var count int
	wg.Add(1)
	go func() {

		var err error
		count, err = db.CountPackages()
		logging.Error(err)

		wg.Done()
	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(count)
	response.RecordsFiltered = strconv.Itoa(count)
	response.Draw = query.Draw

	for _, v := range packages {
		response.AddRow(v.OutputForJSON())
	}

	response.output(w)
}
