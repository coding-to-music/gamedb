package web

import (
	"net/http"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/logging"
)

func statsGenresHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := db.GetConfig(db.ConfGenresUpdated)
	logging.Error(err)

	// Get genres
	genres, err := db.GetAllGenres()
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the genres.", Error: err})
		return
	}

	// Template
	t := statsGenresTemplate{}
	t.Fill(w, r, "Genres")
	t.Genres = genres
	t.Date = config.Value

	err = returnTemplate(w, r, "genres", t)
	logging.Error(err)
}

type statsGenresTemplate struct {
	GlobalTemplate
	Genres []db.Genre
	Date   string
}
