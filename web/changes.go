package web

import (
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/logger"
)

func ChangesHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := changesTemplate{}
	t.Fill(w, r, "Changes")

	returnTemplate(w, r, "changes", t)
}

type changesTemplate struct {
	GlobalTemplate
}

func ChangesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	query.FillFromURL(r.URL.Query())

	var changes []db.Change

	client, ctx, err := db.GetDSClient()
	if err != nil {

		logger.Error(err)

	} else {

		q := datastore.NewQuery(db.KindChange).Limit(100).Order("-change_id")

		q, err = query.SetOrderOffsetDS(q, map[string]string{})
		if err != nil {

			logger.Error(err)

		} else {

			_, err := client.GetAll(ctx, q, &changes)
			logger.Error(err)
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
