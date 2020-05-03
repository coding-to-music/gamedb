package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/webserver/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/tasks"
	"github.com/go-chi/chi"
)

func PublishersRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", publishersHandler)
	return r
}

func publishersHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := tasks.GetTaskConfig(tasks.Publishers{})
	if err != nil {
		err = helpers.IgnoreErrors(err, sql.ErrRecordNotFound)
		log.Err(err, r)
	}

	// Get publishers
	publishers, err := sql.GetAllPublishers()
	if err != nil {
		log.Err(err, r)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the publishers."})
		return
	}

	code := session.GetProductCC(r)
	prices := map[int]string{}
	for _, v := range publishers {
		price, err := v.GetMeanPrice(code)
		log.Err(err, r)
		prices[v.ID] = price
	}

	// Template
	t := statsPublishersTemplate{}
	t.fill(w, r, "Publishers", "Publishers handle marketing and advertising.")
	t.Publishers = publishers
	t.Date = config.Value
	t.Prices = prices

	returnTemplate(w, r, "publishers", t)
}

type statsPublishersTemplate struct {
	GlobalTemplate
	Publishers []sql.Publisher
	Date       string
	Prices     map[int]string
}
