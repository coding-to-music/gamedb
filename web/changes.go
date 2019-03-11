package web

import (
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
	"github.com/go-chi/chi"
)

func changesRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", changesHandler)
	r.Get("/changes.json", changesAjaxHandler)
	r.Get("/{id}", changeHandler)
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

	var limit = 100
	var offset = query.getOffset()

	kinds, err := db.GetBufferRows(db.KindChange, limit, offset)
	if err != nil {
		log.Err(err, r)
		return
	}
	var changes = db.KindsToChanges(kinds)

	if len(changes) < limit {

		limit = limit - len(changes)
		offset = offset - len(changes)

		client, ctx, err := db.GetDSClient()
		if err != nil {

			log.Err(err, r)

		} else {

			q := datastore.NewQuery(db.KindChange).Order("-change_id").Limit(limit).Offset(offset)

			_, err := client.GetAll(ctx, q, &changes)
			log.Err(err, r)
		}
	}

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = "10000"
	response.RecordsFiltered = "10000"
	response.Draw = query.Draw

	for _, v := range changes {
		response.AddRow(v.OutputForJSON())
	}

	response.output(w, r)
}
