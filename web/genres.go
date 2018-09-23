package web

import (
	"net/http"

	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/logger"
)

func StatsGenresHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := db.GetConfig(db.ConfGenresUpdated)
	logger.Error(err)

	// Get genres
	genres, err := db.GetAllGenres()
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
	Genres []db.Genre
	Date   string
}
