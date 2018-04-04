package web

import (
	"net/http"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

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
