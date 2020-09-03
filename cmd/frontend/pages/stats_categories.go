package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/frontend/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi"
)

func CategoriesRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", statsCategoriesHandler)
	return r
}

func statsCategoriesHandler(w http.ResponseWriter, r *http.Request) {

	// Get categories
	categories, err := mysql.GetAllCategories()
	if err != nil {
		log.ErrS(err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the categories."})
		return
	}

	code := session.GetProductCC(r)
	prices := map[int]string{}
	for _, category := range categories {
		price, err := category.GetMeanPrice(code)
		if err != nil {
			log.ErrS(err)
		}
		prices[category.ID] = price
	}

	// Template
	t := statsCategoriesTemplate{}
	t.fill(w, r, "Categories", "Top Steam Categories")
	t.addAssetMark()
	t.Categories = categories
	t.Prices = prices

	returnTemplate(w, r, "stats_categories", t)
}

type statsCategoriesTemplate struct {
	globalTemplate
	Categories []mysql.Category
	Prices     map[int]string
}

func (t statsCategoriesTemplate) includes() []string {
	return []string{"includes/stats_header.gohtml"}
}
