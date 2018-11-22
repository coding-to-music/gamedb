package web

import (
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/logging"
	"github.com/go-chi/chi"
)

func changesRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", ChangesHandler)
	r.Get("/ajax", ChangesAjaxHandler)
	r.Get("/{id}", ChangeHandler)
	return r
}

func ChangesHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := changesTemplate{}
	t.Fill(w, r, "Changes")
	t.Description = "Every time the Steam library gets updated, a change record is created. We use these to keep website information up to date."

	err := returnTemplate(w, r, "changes", t)
	logging.Error(err)
}

type changesTemplate struct {
	GlobalTemplate
}

func ChangesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	query := DataTablesQuery{}
	err := query.FillFromURL(r.URL.Query())
	logging.Error(err)

	var changes []db.Change

	client, ctx, err := db.GetDSClient()
	if err != nil {

		logging.Error(err)

	} else {

		q := datastore.NewQuery(db.KindChange).Limit(100).Order("-change_id")

		q, err = query.SetOrderOffsetDS(q, map[string]string{})
		if err != nil {

			logging.Error(err)

		} else {

			_, err := client.GetAll(ctx, q, &changes)
			logging.Error(err)
		}
	}

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = "10000"
	response.RecordsFiltered = "10000"
	response.Draw = query.Draw

	for _, v := range changes {
		response.AddRow(v.OutputForJSON())
	}

	response.output(w)
}
