package web

import (
	"net/http"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

func StatsDevelopersHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := mysql.GetConfig(mysql.ConfDevelopersUpdated)
	logger.Error(err)

	// Get developers
	developers, err := mysql.GetAllDevelopers()
	if err != nil {
		logger.Error(err)
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
	Developers []mysql.Developer
	Date       string
}
