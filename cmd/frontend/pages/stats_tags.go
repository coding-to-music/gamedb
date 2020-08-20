package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/frontend/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/tasks"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func TagsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", statsTagsHandler)
	return r
}

func statsTagsHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := tasks.GetTaskConfig(tasks.StatsTags{})
	if err != nil {
		err = helpers.IgnoreErrors(err, mysql.ErrRecordNotFound)
		zap.S().Error(err)
	}

	// Get tags
	tags, err := mysql.GetAllTags()
	if err != nil {
		zap.S().Error(err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the tags."})
		return
	}

	code := session.GetProductCC(r)
	prices := map[int]string{}
	for _, v := range tags {
		price, err := v.GetMeanPrice(code)
		if err != nil {
			zap.S().Error(err)
		}
		prices[v.ID] = price
	}

	// Template
	t := statsTagsTemplate{}
	t.fill(w, r, "Tags", "Top Steam tags")
	t.addAssetMark()
	t.Tags = tags
	t.Date = config.Value
	t.Prices = prices

	returnTemplate(w, r, "stats_tags", t)
}

type statsTagsTemplate struct {
	globalTemplate
	Tags   []mysql.Tag
	Date   string
	Prices map[int]string
}
