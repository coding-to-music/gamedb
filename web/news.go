package web

import (
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/logging"
)

func NewsHandler(w http.ResponseWriter, r *http.Request) {

	t := newsTemplate{}
	t.Fill(w, r, "News")
	t.Description = "All the news from all the games, all in one place."

	returnTemplate(w, r, "news", t)
}

type newsTemplate struct {
	GlobalTemplate
}

func NewsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	query.FillFromURL(r.URL.Query())

	var articles []db.News

	client, ctx, err := db.GetDSClient()
	if err != nil {

		logging.Error(err)

	} else {

		q := datastore.NewQuery(db.KindNews).Limit(100)
		q, err = query.SetOrderOffsetDS(q, map[string]string{})
		q = q.Order("-date")
		if err != nil {

			logging.Error(err)

		} else {

			_, err := client.GetAll(ctx, q, &articles)
			logging.Error(err)
		}
	}

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = "10000"
	response.RecordsFiltered = "10000"
	response.Draw = query.Draw

	for _, v := range articles {
		response.AddRow(v.OutputForJSON(r))
	}

	response.output(w)
}

//<tr data-link="{{ .App.GetPath }}#news">
//<td class="img">
//<img class="rounded" src="{{ .App.GetIcon }}" alt="{{ .Article.Title }}">
//<span data-app-id="{{ .App.ID }}">{{ .App.GetName }}</span>
//<td>{{ .Article.Title }}</td>
//<td>{{ .Article.Author }}</td>
//<td><span data-toggle="tooltip" data-placement="top" title="{{ .Article.GetNiceDate }}" data-livestamp="{{ .Article.GetTimestamp }}"></span></td>
//</tr>
