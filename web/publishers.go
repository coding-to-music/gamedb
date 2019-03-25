package web

import (
	"net/http"

	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
	"github.com/gamedb/website/sql"
)

func statsPublishersHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := sql.GetConfig(sql.ConfPublishersUpdated)
	log.Err(err, r)

	// Get publishers
	publishers, err := sql.GetAllPublishers()
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the publishers.", Error: err})
		return
	}

	code := session.GetCountryCode(r)
	prices := map[int]string{}
	for _, v := range publishers {
		price, err := v.GetMeanPrice(code)
		log.Err(err, r)
		prices[v.ID] = price
	}

	// Template
	t := statsPublishersTemplate{}
	t.fill(w, r, "Publishers", "Publishers handle marketing and advertising.")
	t.Publishers = publishers
	t.Date = config.Value
	t.Prices = prices

	err = returnTemplate(w, r, "publishers", t)
	log.Err(err, r)
}

type statsPublishersTemplate struct {
	GlobalTemplate
	Publishers []sql.Publisher
	Date       string
	Prices     map[int]string
}
