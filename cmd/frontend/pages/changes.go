package pages

import (
	"net/http"
	"sync"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

func ChangesRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", changesHandler)
	r.Get("/changes.json", changesAjaxHandler)
	r.Mount("/{id}", ChangeRouter())
	return r
}

func changesHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := changesTemplate{}
	t.fill(w, r, "Changes", "Every time the Steam library gets updated, a change record is created. We use these to keep website information up to date.")

	returnTemplate(w, r, "changes", t)
}

type changesTemplate struct {
	globalTemplate
}

func (t changesTemplate) includes() []string {
	return []string{"includes/changes_header.gohtml"}
}

func changesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, true)

	changes, err := mongo.GetChanges(query.GetOffset64())
	if err != nil {
		log.ErrS(err)
		return
	}

	var wg sync.WaitGroup

	// Get changes
	wg.Add(1)
	var appMap = map[int]string{}
	go func() {

		defer wg.Done()

		var err error
		var appIDs []int

		for _, v := range changes {
			appIDs = append(appIDs, v.Apps...)
		}

		// App map
		apps, err := mongo.GetAppsByID(appIDs, bson.M{"_id": 1, "name": 1})
		if err != nil {
			log.ErrS(err)
		}

		for _, app := range apps {
			appMap[app.ID] = app.GetName()
		}
	}()

	wg.Add(1)
	var packageMap = map[int]string{}
	go func() {

		defer wg.Done()

		var err error
		var packageIDs []int

		for _, v := range changes {
			packageIDs = append(packageIDs, v.Packages...)
		}

		// Package map
		packages, err := mongo.GetPackagesByID(packageIDs, bson.M{"_id": 1, "name": 1})
		if err != nil {
			log.ErrS(err)
		}

		for _, v := range packages {
			packageMap[v.ID] = v.GetName()
		}
	}()

	// Get count
	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		count, err = mongo.CountDocuments(mongo.CollectionChanges, nil, 0)
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, count, count, nil)
	for _, v := range changes {
		response.AddRow(v.OutputForJSON(appMap, packageMap))
	}

	returnJSON(w, r, response)
}
