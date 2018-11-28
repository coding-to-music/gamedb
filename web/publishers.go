package web

import (
	"net/http"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
)

func statsPublishersHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := db.GetConfig(db.ConfPublishersUpdated)
	log.Log(err)

	// Get publishers
	publishers, err := db.GetAllPublishers()
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the publishers.", Error: err})
		return
	}

	// Template
	t := statsPublishersTemplate{}
	t.Fill(w, r, "Publishers")
	t.Publishers = publishers
	t.Date = config.Value

	err = returnTemplate(w, r, "publishers", t)
	log.Log(err)
}

type statsPublishersTemplate struct {
	GlobalTemplate
	Publishers []db.Publisher
	Date       string
}
