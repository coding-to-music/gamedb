package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/frontend/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi"
)

func GenresRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", genresHandler)
	return r
}

func genresHandler(w http.ResponseWriter, r *http.Request) {

	// Get genres
	genres, err := mysql.GetAllGenres(false)
	if err != nil {
		log.ErrS(err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the genres."})
		return
	}

	prices := map[int]string{}
	for _, v := range genres {
		price, err := v.GetMeanPrice(session.GetProductCC(r))
		if err != nil {
			log.ErrS(err)
		}
		prices[v.ID] = price
	}

	// Template
	t := statsGenresTemplate{}
	t.fill(w, r, "Genres", "All Steam genres")
	t.addAssetMark()
	t.Genres = genres
	t.Prices = prices

	returnTemplate(w, r, "stats_genres", t)
}

type statsGenresTemplate struct {
	globalTemplate
	Genres []mysql.Genre
	Prices map[int]string
}

func (t statsGenresTemplate) includes() []string {
	return []string{"includes/stats_header.gohtml"}
}
