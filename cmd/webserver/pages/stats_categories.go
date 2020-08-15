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

func CategoriesRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", statsCategoriesHandler)
	return r
}

func statsCategoriesHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := tasks.GetTaskConfig(tasks.StatsCategories{})
	if err != nil {
		err = helpers.IgnoreErrors(err, mysql.ErrRecordNotFound)
		zap.S().Error(err)
	}

	// Get categories
	categories, err := mysql.GetAllCategories()
	if err != nil {
		zap.S().Error(err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the categories."})
		return
	}

	code := session.GetProductCC(r)
	prices := map[int]string{}
	for _, category := range categories {
		price, err := category.GetMeanPrice(code)
		zap.S().Error(err)
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
	globalTemplate
	Categories []mysql.Category
	Date       string
	Prices     map[int]string
}
