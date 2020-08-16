package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/tasks"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func DevelopersRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", developersHandler)
	return r
}

func developersHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := tasks.GetTaskConfig(tasks.StatsDevelopers{})
	if err != nil {
		err = helpers.IgnoreErrors(err, mysql.ErrRecordNotFound)
		zap.S().Error(err)
	}

	// Get developers
	developers, err := mysql.GetAllDevelopers([]string{})
	if err != nil {
		zap.S().Error(err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the developers."})
		return
	}

	prices := map[int]string{}
	for _, v := range developers {
		price, err := v.GetMeanPrice(session.GetProductCC(r))
		if err != nil {
			zap.S().Error(err)
		}
		prices[v.ID] = price
	}

	// Template
	t := statsDevelopersTemplate{}
	t.fill(w, r, "Developers", "All the software developers that create Steam content.")
	t.addAssetMark()
	t.Developers = developers
	t.Date = config.Value
	t.Prices = prices

	returnTemplate(w, r, "stats_developers", t)
}

type statsDevelopersTemplate struct {
	globalTemplate
	Developers []mysql.Developer
	Date       string
	Prices     map[int]string
}
