package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/frontend/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi"
)

func TagsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", statsTagsHandler)
	return r
}

func statsTagsHandler(w http.ResponseWriter, r *http.Request) {

	// Get tags
	tags, err := mysql.GetAllTags()
	if err != nil {
		log.ErrS(err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the tags."})
		return
	}

	code := session.GetProductCC(r)
	prices := map[int]string{}
	for _, v := range tags {
		price, err := v.GetMeanPrice(code)
		if err != nil {
			log.ErrS(err)
		}
		prices[v.ID] = price
	}

	// Template
	t := statsTagsTemplate{}
	t.fill(w, r, "Tags", "Top Steam tags")
	t.addAssetMark()
	t.Tags = tags
	t.Prices = prices

	returnTemplate(w, r, "stats_tags", t)
}

type statsTagsTemplate struct {
	globalTemplate
	Tags   []mysql.Tag
	Prices map[int]string
}

func (t statsTagsTemplate) includes() []string {
	return []string{"includes/stats_header.gohtml"}
}
