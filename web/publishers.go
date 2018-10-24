package web

import (
	"net/http"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/logging"
)

func StatsPublishersHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := db.GetConfig(db.ConfPublishersUpdated)
	logging.Error(err)

	// Get publishers
	publishers, err := db.GetAllPublishers()
	if err != nil {
		logging.Error(err)
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
	Publishers []db.Publisher
	Date       string
}
