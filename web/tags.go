package web

import (
	"net/http"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

func StatsTagsHandler(w http.ResponseWriter, r *http.Request) {

	tags, err := mysql.GetAllTags()
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, "Error getting tags")
		return
	}

	// Template
	template := statsTagsTemplate{}
	template.Fill(w, r, "Tags")
	template.Tags = tags

	returnTemplate(w, r, "tags", template)
	return
}

type statsTagsTemplate struct {
	GlobalTemplate
	Tags []mysql.Tag
}
