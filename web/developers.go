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
	log.Log(err)

	// Get developers
	developers, err := db.GetAllDevelopers()
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the developers.", Error: err})
		return
	}

	code := session.GetCountryCode(r)
	prices := map[int]string{}
	for _, v := range developers {
		price, err := v.GetMeanPrice(code)
		log.Log(err)
		prices[v.ID] = price
	}

	// Template
	t := statsDevelopersTemplate{}
	t.Fill(w, r, "Developers", "All the software developers that create Steam content.")
	t.Developers = developers
	t.Date = config.Value
	t.Prices = prices

	err = returnTemplate(w, r, "developers", t)
	log.Log(err)
}

type statsDevelopersTemplate struct {
	GlobalTemplate
	Developers []db.Developer
	Date       string
	Prices     map[int]string
}
