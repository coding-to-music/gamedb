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
	r.Get("/ajax", changesAjaxHandler)
	r.Get("/{id}", changeHandler)
	return r
}

func changesHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := changesTemplate{}
	t.Fill(w, r, "Changes", "Every time the Steam library gets updated, a change record is created. We use these to keep website information up to date.")

	err := returnTemplate(w, r, "changes", t)
	log.Log(err)
}

type changesTemplate struct {
	GlobalTemplate
}

func changesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	query := DataTablesQuery{}
	err := query.FillFromURL(r.URL.Query())
	log.Log(err)

	var changes []db.Change

	client, ctx, err := db.GetDSClient()
	if err != nil {

		log.Log(err)

	} else {

		q := datastore.NewQuery(db.KindChange).Limit(100).Order("-change_id")

		q, err = query.SetOrderOffsetDS(q, map[string]string{})
		if err != nil {

			log.Log(err)

		} else {

			_, err := client.GetAll(ctx, q, &changes)
			log.Log(err)
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
