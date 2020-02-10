package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/tasks"
	"github.com/go-chi/chi"
)

func GenresRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", genresHandler)
	return r
}

func genresHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := tasks.GetTaskConfig(tasks.Genres{})
	if err != nil {
		err = helpers.IgnoreErrors(err, sql.ErrRecordNotFound)
		log.Err(err, r)
	}

	// Get genres
	genres, err := sql.GetAllGenres(false)
	if err != nil {
		log.Err(err, r)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the genres."})
		return
	}

	prices := map[int]string{}
	for _, v := range genres {
		price, err := v.GetMeanPrice(helpers.GetProductCC(r))
		log.Err(err, r)
		prices[v.ID] = price
	}

	// Template
	t := statsGenresTemplate{}
	t.fill(w, r, "Genres", "All Steam genres")
	t.Genres = genres
	t.Date = config.Value
	t.Prices = prices

	returnTemplate(w, r, "genres", t)
}

type statsGenresTemplate struct {
	GlobalTemplate
	Genres []sql.Genre
	Date   string
	Prices map[int]string
}
