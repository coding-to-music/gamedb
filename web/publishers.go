package web

import (
	"net/http"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

func StatsPublishersHandler(w http.ResponseWriter, r *http.Request) {

	// Get publishers
	publishers, err := mysql.GetAllPublishers()
	if err != nil {
		logger.Error(err)
	}

	// Template
	template := statsPublishersTemplate{}
	template.Fill(w, r, "Publishers")
	template.Publishers = publishers

	returnTemplate(w, r, "publishers", template)
	return
}

type statsPublishersTemplate struct {
	GlobalTemplate
	Publishers []mysql.Publisher
}
