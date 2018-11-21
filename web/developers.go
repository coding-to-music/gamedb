package web

import (
	"net/http"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/logging"
)

func statsDevelopersHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := db.GetConfig(db.ConfDevelopersUpdated)
	logging.Error(err)

	// Get developers
	developers, err := db.GetAllDevelopers()
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the developers.", Error: err})
		return
	}

	// Template
	t := statsDevelopersTemplate{}
	t.Fill(w, r, "Developers")
	t.Developers = developers
	t.Date = config.Value

	err := returnTemplate(w, r, "developers", t)
	logging.Error(err)
}

type statsDevelopersTemplate struct {
	GlobalTemplate
	Developers []db.Developer
	Date       string
}
