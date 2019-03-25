package web

import (
	"net/http"

	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
	"github.com/gamedb/website/sql"
)

func statsGenresHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := sql.GetConfig(sql.ConfGenresUpdated)
	log.Err(err, r)

	// Get genres
	genres, err := sql.GetAllGenres()
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the genres.", Error: err})
		return
	}

	code := session.GetCountryCode(r)
	prices := map[int]string{}
	for _, v := range genres {
		price, err := v.GetMeanPrice(code)
		log.Err(err, r)
		prices[v.ID] = price
	}

	// Template
	t := statsGenresTemplate{}
	t.fill(w, r, "Genres", "")
	t.Genres = genres
	t.Date = config.Value
	t.Prices = prices

	err = returnTemplate(w, r, "genres", t)
	log.Err(err, r)
}

type statsGenresTemplate struct {
	GlobalTemplate
	Genres []sql.Genre
	Date   string
	Prices map[int]string
}
