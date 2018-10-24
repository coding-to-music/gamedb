package web

import (
	"net/http"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/logging"
)

func StatsDevelopersHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := db.GetConfig(db.ConfDevelopersUpdated)
	logging.Error(err)

	// Get developers
	developers, err := db.GetAllDevelopers()
	if err != nil {
		logging.Error(err)
		returnErrorTemplate(w, r, 500, "Error getting developers")
		return
	}

	// Template
	t := statsDevelopersTemplate{}
	t.Fill(w, r, "Developers")
	t.Developers = developers
	t.Date = config.Value

	returnTemplate(w, r, "developers", t)
	return
}

type statsDevelopersTemplate struct {
	GlobalTemplate
	Developers []db.Developer
	Date       string
}
