package pages

import (
	"html/template"
	"net/http"
	"sync"

	"github.com/gamedb/gamedb/cmd/webserver/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

func PackagesRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", packagesHandler)
	r.Get("/packages.json", packagesAjaxHandler)
	r.Mount("/{id}", PackageRouter())
	return r
}

func packagesHandler(w http.ResponseWriter, r *http.Request) {

	total, err := mongo.CountDocuments(mongo.CollectionPackages, nil, 0)
	if err != nil {
		log.Err(err, r)
	}

	// Template
	t := packagesTemplate{}
	t.fill(w, r, "Packages", "The last "+template.HTML(helpers.ShortHandNumber(total))+" packages to be updated.")

	returnTemplate(w, r, "packages", t)
}

type packagesTemplate struct {
	GlobalTemplate
}

func packagesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, false)

	//
	var code = helpers.GetProductCC(r)
	var wg sync.WaitGroup

	// Get apps
	var packages []mongo.Package

	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		var err error
		var projection = bson.M{"id": 1, "name": 1, "apps_count": 1, "change_number_date": 1, "prices": 1, "icon": 1}
		var sortCols = map[string]string{
			"1": "prices." + string(code) + ".final",
			"2": "prices." + string(code) + ".discount_percent",
			"3": "apps_count",
			"4": "change_number_date",
		}

		packages, err = mongo.GetPackages(query.GetOffset64(), 100, query.GetOrderMongo(sortCols), nil, projection, nil)
		if err != nil {
			log.Err(err, r)
			return
		}
	}(r)

	// Get total
	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountDocuments(mongo.CollectionPackages, nil, 0)
		log.Err(err, r)

	}()

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, count, count)
	for _, v := range packages {
		response.AddRow(v.OutputForJSON(code))
	}

	returnJSON(w, r, response)
}
