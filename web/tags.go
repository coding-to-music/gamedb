package web

import (
	"net/http"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
)

func statsTagsHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := db.GetConfig(db.ConfTagsUpdated)
	log.Err(err, r)

	// Get tags
	tags, err := db.GetAllTags()
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the tags.", Error: err})
		return
	}

	code := session.GetCountryCode(r)
	prices := map[int]string{}
	for _, v := range tags {
		price, err := v.GetMeanPrice(code)
		log.Err(err, r)
		prices[v.ID] = price
	}

	// Template
	t := statsTagsTemplate{}
	t.fill(w, r, "Tags", "")
	t.Tags = tags
	t.Date = config.Value
	t.Prices = prices

	err = returnTemplate(w, r, "tags", t)
	log.Err(err, r)
}

type statsTagsTemplate struct {
	GlobalTemplate
	Tags   []db.Tag
	Date   string
	Prices map[int]string
}
