package web

import (
	"net/http"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
)

func statsPublishersHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := db.GetConfig(db.ConfPublishersUpdated)
	log.Err(err)

	// Get publishers
	publishers, err := db.GetAllPublishers()
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the publishers.", Error: err})
		return
	}

	code := session.GetCountryCode(r)
	prices := map[int]string{}
	for _, v := range publishers {
		price, err := v.GetMeanPrice(code)
		log.Err(err)
		prices[v.ID] = price
	}

	// Template
	t := statsPublishersTemplate{}
	t.Fill(w, r, "Publishers", "Publishers handle marketing and advertising.")
	t.Publishers = publishers
	t.Date = config.Value
	t.Prices = prices

	err = returnTemplate(w, r, "publishers", t)
	log.Err(err)
}

type statsPublishersTemplate struct {
	GlobalTemplate
	Publishers []db.Publisher
	Date       string
	Prices     map[int]string
}
