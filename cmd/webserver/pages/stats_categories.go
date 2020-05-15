package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/tasks"
	"github.com/go-chi/chi"
)

func CategoriesRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", statsCategoriesHandler)
	return r
}

func statsCategoriesHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := tasks.GetTaskConfig(tasks.StatsCategories{})
	if err != nil {
		err = helpers.IgnoreErrors(err, sql.ErrRecordNotFound)
		log.Err(err, r)
	}

	// Get categories
	categories, err := sql.GetAllCategories()
	if err != nil {
		log.Err(err, r)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the categories."})
		return
	}

	code := session.GetProductCC(r)
	prices := map[int]string{}
	for _, category := range categories {
		price, err := category.GetMeanPrice(code)
		log.Err(err, r)
		prices[category.ID] = price
	}

	// Template
	t := statsCategoriesTemplate{}
	t.fill(w, r, "Categories", "Top Steam Categories")
	t.addAssetMark()
	t.Categories = categories
	t.Date = config.Value
	t.Prices = prices

	returnTemplate(w, r, "stats_categories", t)
}

type statsCategoriesTemplate struct {
	GlobalTemplate
	Categories []sql.Category
	Date       string
	Prices     map[int]string
}
