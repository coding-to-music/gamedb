package pages

import (
	"html/template"
	"net/http"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
)

func PackagesRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", packagesHandler)
	r.Get("/packages.json", packagesAjaxHandler)
	r.Mount("/{id}", PackageRouter())
	return r
}

func packagesHandler(w http.ResponseWriter, r *http.Request) {

	total, err := sql.CountPackages()
	log.Err(err, r)

	// Template
	t := packagesTemplate{}
	t.fill(w, r, "Packages", "The last "+template.HTML(humanize.Comma(int64(total)))+" packages to be updated.")

	err = returnTemplate(w, r, "packages", t)
	log.Err(err, r)
}

type packagesTemplate struct {
	GlobalTemplate
}

func packagesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	//
	var code = helpers.GetProductCC(r)
	var wg sync.WaitGroup

	// Get apps
	var packages []sql.Package

	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		gorm, err := sql.GetMySQLClient()
		if err != nil {

			log.Err(err, r)
			return
		}

		gorm = gorm.Model(&sql.Package{})
		gorm = gorm.Select([]string{"id", "name", "apps_count", "change_number_date", "prices", "coming_soon", "icon"})

		gorm = query.setOrderOffsetGorm(gorm, code, map[string]string{
			"2": "apps_count",
			"3": "price",
			"4": "change_number_date",
		})

		gorm = gorm.Limit(100)
		gorm = gorm.Find(&packages)

		log.Err(gorm.Error)

	}(r)

	// Get total
	var count int
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = sql.CountPackages()
		log.Err(err, r)

	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = int64(count)
	response.RecordsFiltered = int64(count)
	response.Draw = query.Draw

	for _, v := range packages {
		response.AddRow(v.OutputForJSON(code))
	}

	response.output(w, r)
}
