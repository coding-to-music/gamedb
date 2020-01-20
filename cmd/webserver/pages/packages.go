package pages

import (
	"html/template"
	"net/http"
	"sync"

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
	t.fill(w, r, "Packages", "The last "+template.HTML(helpers.ShortHandNumber(int64(total)))+" packages to be updated.")

	returnTemplate(w, r, "packages", t)
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

		db, err := sql.GetMySQLClient()
		if err != nil {
			log.Err(err, r)
			return
		}

		db = db.Model(&sql.Package{})
		db = db.Select([]string{"id", "name", "apps_count", "change_number_date", "prices", "icon"})

		sortCols := map[string]string{
			"1": "JSON_EXTRACT(prices, \"$." + string(code) + ".final\")",
			"2": "JSON_EXTRACT(prices, \"$." + string(code) + ".discount_percent\")",
			"3": "apps_count",
			"4": "change_number_date",
		}
		db = query.setOrderOffsetGorm(db, sortCols, "4")

		db = db.Limit(100)
		db = db.Find(&packages)

		log.Err(db.Error)

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

	response := DataTablesResponse{}
	response.RecordsTotal = int64(count)
	response.RecordsFiltered = int64(count)
	response.Draw = query.Draw

	for _, v := range packages {
		response.AddRow(v.OutputForJSON(code))
	}

	response.output(w, r)
}
