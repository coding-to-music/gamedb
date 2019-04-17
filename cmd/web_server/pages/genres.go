package pages

import (
	"net/http"
	"time"

	"github.com/gamedb/website/pkg"
	"github.com/go-chi/chi"
)

func genresRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", genresHandler)
	return r
}

func genresHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*24)

	// Get config
	config, err := pkg.GetConfig(pkg.ConfGenresUpdated)
	log.Err(err, r)

	// Get genres
	genres, err := pkg.GetAllGenres(false)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the genres.", Error: err})
		return
	}

	code := pkg.GetCountryCode(r)
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
