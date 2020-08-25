package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/frontend/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
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
	config, err := tasks.GetTaskConfig(tasks.TasksPublishers{})
	if err != nil {
		err = helpers.IgnoreErrors(err, mysql.ErrRecordNotFound)
		if err != nil {
			log.ErrS(err)
		}
	}

	// Get publishers
	publishers, err := mysql.GetAllPublishers()
	if err != nil {
		log.ErrS(err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the publishers."})
		return
	}

	code := session.GetProductCC(r)
	prices := map[int]string{}
	for _, v := range publishers {
		price, err := v.GetMeanPrice(code)
		if err != nil {
			log.ErrS(err)
		}
		prices[v.ID] = price
	}

	// Template
	t := statsPublishersTemplate{}
	t.fill(w, r, "Publishers", "Publishers handle marketing and advertising.")
	t.addAssetMark()
	t.Publishers = publishers
	t.Date = config.Value
	t.Prices = prices

	returnTemplate(w, r, "stats_publishers", t)
}

type statsPublishersTemplate struct {
	globalTemplate
	Publishers []mysql.Publisher
	Date       string
	Prices     map[int]string
}

func (t statsPublishersTemplate) includes() []string {
	return []string{"includes/stats_header.gohtml"}
}
