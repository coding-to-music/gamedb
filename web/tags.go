package web

import (
	"net/http"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/logging"
)

func statsTagsHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := db.GetConfig(db.ConfTagsUpdated)
	logging.Error(err)

	// Get tags
	tags, err := db.GetAllTags()
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the tags.", Error: err})
		return
	}

	// Template
	t := statsTagsTemplate{}
	t.Fill(w, r, "Tags")
	t.Tags = tags
	t.Date = config.Value

	returnTemplate(w, r, "tags", t)
}

type statsTagsTemplate struct {
	GlobalTemplate
	Tags []db.Tag
	Date string
}
