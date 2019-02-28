package web

import (
	"net/http"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
)

func statsDevelopersHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := db.GetConfig(db.ConfDevelopersUpdated)
	log.Err(err, r)

	// Get developers
	developers, err := db.GetAllDevelopers([]string{})
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the developers.", Error: err})
		return
	}

	code := session.GetCountryCode(r)
	prices := map[int]string{}
	for _, v := range developers {
		price, err := v.GetMeanPrice(code)
		log.Err(err, r)
		prices[v.ID] = price
	}

	// Template
	t := statsDevelopersTemplate{}
	t.Fill(w, r, "Developers", "All the software developers that create Steam content.")
	t.Developers = developers
	t.Date = config.Value
	t.Prices = prices

	err = returnTemplate(w, r, "developers", t)
	log.Err(err, r)
}

type statsDevelopersTemplate struct {
	GlobalTemplate
	Developers []db.Developer
	Date       string
	Prices     map[int]string
}
