package pages

import (
	"html/template"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/website/pkg"
	"github.com/go-chi/chi"
)

func packagesRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", packagesHandler)
	r.Get("/packages.json", packagesAjaxHandler)
	r.Mount("/{id}", packageRouter())
	return r
}

func packagesHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*24)

	total, err := pkg.CountPackages()
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

	ret := setAllowedQueries(w, r, []string{"draw", "order[0][column]", "order[0][dir]", "start"})
	if ret {
		return
	}

	setCacheHeaders(w, 0)

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	//
	var code = pkg.GetCountryCode(r)
	var wg sync.WaitGroup

	// Get apps
	var packages []pkg.Package

	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		gorm, err := sql.GetMySQLClient()
		if err != nil {

			log.Err(err, r)
			return
		}

		gorm = gorm.Model(&pkg.Package{})
		gorm = gorm.Select([]string{"id", "name", "apps_count", "change_number_date", "prices", "coming_soon", "icon"})

		gorm = query.setOrderOffsetGorm(gorm, code, map[string]string{
			"0": "name",
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
		count, err = pkg.CountPackages()
		log.Err(err, r)

	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(count)
	response.RecordsFiltered = strconv.Itoa(count)
	response.Draw = query.Draw

	for _, v := range packages {
		response.AddRow(v.OutputForJSON(code))
	}

	response.output(w, r)
}
