package web

import (
	"net/http"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

func StatsGenresHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := mysql.GetConfig(mysql.ConfGenresUpdated)
	logger.Error(err)

	// Get genres
	genres, err := mysql.GetAllGenres()
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, "Error getting genres")
		return
	}

	// Template
	t := statsGenresTemplate{}
	t.Fill(w, r, "Genres")
	t.Genres = genres
	t.Date = config.Value

	returnTemplate(w, r, "genres", t)
	return
}

type statsGenresTemplate struct {
	GlobalTemplate
	Genres []mysql.Genre
	Date   string
}
