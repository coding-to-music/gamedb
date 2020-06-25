package pages

import (
	"net/http"
	"sync"

	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi"
)

func BundlesRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", bundlesHandler)
	r.Get("/bundles.json", bundlesAjaxHandler)
	r.Mount("/{id}", BundleRouter())
	return r
}

func bundlesHandler(w http.ResponseWriter, r *http.Request) {

	t := bundlesTemplate{}
	t.fill(w, r, "Bundles", "All the bundles on Steam")

	returnTemplate(w, r, "bundles", t)
}

type bundlesTemplate struct {
	GlobalTemplate
}

func bundlesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, false)

	//
	var wg sync.WaitGroup

	// Get apps
	var bundles []mysql.Bundle

	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		gorm, err := mysql.GetMySQLClient()
		if err != nil {

			log.Err(err, r)
			return
		}

		gorm = gorm.Model(&mysql.Bundle{})
		gorm = gorm.Select([]string{"id", "name", "updated_at", "discount", "highest_discount", "app_ids", "package_ids"})
		gorm = gorm.Limit(100)

		sortCols := map[string]string{
			"1": "discount",
			"2": "JSON_LENGTH(app_ids)",
			"3": "JSON_LENGTH(package_ids)",
			"4": "updated_at",
		}
		gorm = query.SetOrderOffsetGorm(gorm, sortCols)

		gorm = gorm.Find(&bundles)

		log.Err(gorm.Error, r)
	}(r)

	// Get total
	var count int
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mysql.CountBundles()
		log.Err(err, r)

	}()

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, int64(count), int64(count), nil)
	for _, v := range bundles {
		response.AddRow(v.OutputForJSON())
	}

	returnJSON(w, r, response)
}
