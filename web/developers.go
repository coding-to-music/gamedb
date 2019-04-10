package web

import (
	"net/http"
	"time"

	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
	"github.com/gamedb/website/sql"
)

func statsDevelopersHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*24)

	// Get config
	config, err := sql.GetConfig(sql.ConfDevelopersUpdated)
	log.Err(err, r)

	// Get developers
	developers, err := sql.GetAllDevelopers([]string{})
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
	t.fill(w, r, "Developers", "All the software developers that create Steam content.")
	t.Developers = developers
	t.Date = config.Value
	t.Prices = prices

	err = returnTemplate(w, r, "developers", t)
	log.Err(err, r)
}

type statsDevelopersTemplate struct {
	GlobalTemplate
	Developers []sql.Developer
	Date       string
	Prices     map[int]string
}
