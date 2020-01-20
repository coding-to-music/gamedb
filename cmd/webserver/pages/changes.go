package pages

import (
	"net/http"
	"sync"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
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
	GlobalTemplate
}

func changesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	if err != nil {
		log.Err(err, r)
		return
	}

	query.limit(r)

	var wg sync.WaitGroup

	// Get changes
	var changes []mongo.Change
	var appMap = map[int]string{}
	var packageMap = map[int]string{}
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		changes, err = mongo.GetChanges(query.getOffset64())
		if err != nil {
			log.Err(err, r)
			return
		}

		var appIDs []int
		var packageIDs []int

		for _, v := range changes {
			appIDs = append(appIDs, v.Apps...)
			packageIDs = append(packageIDs, v.Packages...)
		}

		// App map
		apps, err := sql.GetAppsByID(appIDs, []string{"id", "name"})
		log.Err(err)

		for _, v := range apps {
			appMap[v.ID] = v.GetName()
		}

		// Package map
		packages, err := sql.GetPackages(packageIDs, []string{"id", "name"})
		log.Err(err)

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
			log.Err(err, r)
		}
	}()

	wg.Wait()

	response := DataTablesResponse{}
	response.RecordsTotal = count
	response.RecordsFiltered = count
	response.Draw = query.Draw
	response.limit(r)

	for _, v := range changes {
		response.AddRow(v.OutputForJSON(appMap, packageMap))
	}

	response.output(w, r)
}
