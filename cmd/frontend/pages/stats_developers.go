package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/frontend/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi"
)

func DevelopersRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", developersHandler)
	return r
}

func developersHandler(w http.ResponseWriter, r *http.Request) {

	// Get developers
	developers, err := mysql.GetAllDevelopers([]string{})
	if err != nil {
		log.ErrS(err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the developers."})
		return
	}

	prices := map[int]string{}
	for _, v := range developers {
		price, err := v.GetMeanPrice(session.GetProductCC(r))
		if err != nil {
			log.ErrS(err)
		}
		prices[v.ID] = price
	}

	// Template
	t := statsDevelopersTemplate{}
	t.fill(w, r, "Developers", "All the software developers that create Steam content.")
	t.addAssetMark()
	t.Developers = developers
	t.Prices = prices

	returnTemplate(w, r, "stats_developers", t)
}

type statsDevelopersTemplate struct {
	globalTemplate
	Developers []mysql.Developer
	Prices     map[int]string
}

func (t statsDevelopersTemplate) includes() []string {
	return []string{"includes/stats_header.gohtml"}
}
