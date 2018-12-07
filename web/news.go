package web

import (
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
)

func newsHandler(w http.ResponseWriter, r *http.Request) {

	t := newsTemplate{}
	t.Fill(w, r, "News", "All the news from all the games, all in one place.")

	err := returnTemplate(w, r, "news", t)
	log.Log(err)
}

type newsTemplate struct {
	GlobalTemplate
}

func newsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	query := DataTablesQuery{}
	err := query.FillFromURL(r.URL.Query())
	log.Log(err)

	var articles []db.News

	client, ctx, err := db.GetDSClient()
	if err != nil {

		log.Log(err)

	} else {

		q := datastore.NewQuery(db.KindNews).Limit(100)
		q, err = query.SetOrderOffsetDS(q, map[string]string{})
		q = q.Order("-date")
		if err != nil {

			log.Log(err)

		} else {

			_, err := client.GetAll(ctx, q, &articles)
			log.Log(err)

			for k, v := range articles {
				articles[k].Contents = helpers.BBCodeCompiler.Compile(v.Contents)
			}
		}
	}

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = "10000"
	response.RecordsFiltered = "10000"
	response.Draw = query.Draw

	for _, v := range articles {
		response.AddRow(v.OutputForJSON(r))
	}

	response.output(w, r)
}
