package web

import (
	"net/http"
	"sort"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

func StatsHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	template := statsTemplate{}
	template.Fill(r, "Stats")

	returnTemplate(w, r, "stats", template)

}

type statsTemplate struct {
	GlobalTemplate
}

func StatsGenresHandler(w http.ResponseWriter, r *http.Request) {

	genres, err := mysql.GetAllGenres()
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, "Error getting genres")
		return
	}

	// Template
	template := statsGenresTemplate{}
	template.Fill(r, "Genres")
	template.Genres = genres

	returnTemplate(w, r, "genres", template)
	return
}

type statsGenresTemplate struct {
	GlobalTemplate
	Genres []mysql.Genre
}

func StatsTagsHandler(w http.ResponseWriter, r *http.Request) {

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
	template := statsTagsTemplate{}
	template.Fill(r, "Tags")
	template.Tags = tags

	returnTemplate(w, r, "tags", template)
	return
}

type statsTagsTemplate struct {
	GlobalTemplate
	Tags []mysql.Tag
}

func StatsDevelopersHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	template := statsDevelopersTemplate{}
	template.Fill(r, "Developers")

	returnTemplate(w, r, "developers", template)
	return
}

type statsDevelopersTemplate struct {
	GlobalTemplate
	Developers []int
}

func StatsPublishersHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	template := statsPublishersTemplate{}
	template.Fill(r, "Publishers")

	returnTemplate(w, r, "publishers", template)
	return
}

type statsPublishersTemplate struct {
	GlobalTemplate
	Publishers []int
}
