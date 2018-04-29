package web

import (
	"net/http"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

func StatsPublishersHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := mysql.GetConfig(mysql.ConfPublishersUpdated)
	logger.Error(err)

	// Get publishers
	publishers, err := mysql.GetAllPublishers()
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, "Error getting publishers")
		return
	}

	// Template
	t := statsPublishersTemplate{}
	t.Fill(w, r, "Publishers")
	t.Publishers = publishers
	t.Date = config.Value

	returnTemplate(w, r, "publishers", t)
	return
}

type statsPublishersTemplate struct {
	GlobalTemplate
	Publishers []mysql.Publisher
	Date       string
}
