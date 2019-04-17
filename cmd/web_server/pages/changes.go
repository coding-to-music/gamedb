package pages

import (
	"net/http"

	"github.com/gamedb/website/pkg"
	"github.com/go-chi/chi"
)

func changesRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", changesHandler)
	r.Get("/changes.json", changesAjaxHandler)
	r.Mount("/{id}", changeRouter())
	return r
}

func changesHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := changesTemplate{}
	t.fill(w, r, "Changes", "Every time the Steam library gets updated, a change record is created. We use these to keep website information up to date.")

	err := returnTemplate(w, r, "changes", t)
	log.Err(err, r)
}

type changesTemplate struct {
	GlobalTemplate
}

func changesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	setCacheHeaders(w, 0)

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	if err != nil {
		log.Err(err, r)
		return
	}

	changes, err := pkg.GetChanges(query.getOffset64())
	if err != nil {
		log.Err(err, r)
		return
	}

	var appIDs []int
	var packageIDs []int
	var appMap = map[int]string{}
	var packageMap = map[int]string{}

	for _, v := range changes {
		appIDs = append(appIDs, v.Apps...)
		packageIDs = append(packageIDs, v.Packages...)
	}

	apps, err := pkg.GetAppsByID(appIDs, []string{"id", "name"})
	log.Err(err)

	for _, v := range apps {
		appMap[v.ID] = v.GetName()
	}

	packages, err := pkg.GetPackages(packageIDs, []string{"id", "name"})
	log.Err(err)

	for _, v := range packages {
		packageMap[v.ID] = v.GetName()
	}

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = "10000"
	response.RecordsFiltered = "10000"
	response.Draw = query.Draw

	for _, v := range changes {
		response.AddRow(v.OutputForJSON(appMap, packageMap))
	}

	response.output(w, r)
}
