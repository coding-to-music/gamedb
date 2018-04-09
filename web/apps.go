package web

import (
	"net/http"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

func AppsHandler(w http.ResponseWriter, r *http.Request) {

	// Get apps
	apps, err := mysql.SearchApps(r.URL.Query(), 96, "id DESC", []string{})
	if err != nil {
		logger.Error(err)
	}

	// Get apps count
	count, err := mysql.CountApps()
	if err != nil {
		logger.Error(err)
	}

	// Template
	template := appsTemplate{}
	template.Fill(r, "Games")
	template.Apps = apps
	template.Count = count

	returnTemplate(w, r, "apps", template)
}

type appsTemplate struct {
	GlobalTemplate
	Apps  []mysql.App
	Count int
}
