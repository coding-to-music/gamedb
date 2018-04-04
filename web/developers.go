package web

import (
	"net/http"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

func StatsDevelopersHandler(w http.ResponseWriter, r *http.Request) {

	// Get developers
	developers, err := mysql.GetAllDevelopers()
	if err != nil {
		logger.Error(err)
	}

	// Template
	template := statsDevelopersTemplate{}
	template.Fill(r, "Developers")
	template.Developers = developers

	returnTemplate(w, r, "developers", template)
	return
}

type statsDevelopersTemplate struct {
	GlobalTemplate
	Developers []mysql.Developer
}
