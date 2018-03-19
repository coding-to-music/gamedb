package web

import (
	"net/http"
	"sort"

	"github.com/Jleagle/go-helpers/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

func TagsHandler(w http.ResponseWriter, r *http.Request) {

	tags, err := mysql.GetAllTags()
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, "Error getting tags")
		return
	}

	// Sort friends by level desc
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Apps > tags[j].Apps
	})

	// Template
	template := tagsTemplate{}
	template.Fill(r, "Tags")
	template.Tags = tags

	returnTemplate(w, r, "tags", template)
	return
}

type tagsTemplate struct {
	GlobalTemplate
	Tags []mysql.Tag
}
