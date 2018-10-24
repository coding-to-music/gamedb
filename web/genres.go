package web

import (
	"net/http"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/logging"
)

func StatsGenresHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := db.GetConfig(db.ConfGenresUpdated)
	logging.Error(err)

	// Get genres
	genres, err := db.GetAllGenres()
	if err != nil {
		logging.Error(err)
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
