package web

import (
	"net/http"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

func StatsTagsHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := mysql.GetConfig(mysql.ConfTagsUpdated)
	logger.Error(err)

	// Get tags
	tags, err := mysql.GetAllTags()
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, "Error getting tags")
		return
	}

	// Template
	t := statsTagsTemplate{}
	t.Fill(w, r, "Tags")
	t.Tags = tags
	t.Date = config.Value

	returnTemplate(w, r, "tags", t)
	return
}

type statsTagsTemplate struct {
	GlobalTemplate
	Tags []mysql.Tag
	Date string
}
