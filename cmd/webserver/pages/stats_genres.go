package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/tasks"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func GenresRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", genresHandler)
	return r
}

func genresHandler(w http.ResponseWriter, r *http.Request) {

	// Get config
	config, err := tasks.GetTaskConfig(tasks.TasksGenres{})
	if err != nil {
		err = helpers.IgnoreErrors(err, mysql.ErrRecordNotFound)
		zap.S().Error(err)
	}

	// Get genres
	genres, err := mysql.GetAllGenres(false)
	if err != nil {
		zap.S().Error(err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the genres."})
		return
	}

	prices := map[int]string{}
	for _, v := range genres {
		price, err := v.GetMeanPrice(session.GetProductCC(r))
		if err != nil {
			zap.S().Error(err)
		}
		prices[v.ID] = price
	}

	// Template
	t := statsGenresTemplate{}
	t.fill(w, r, "Genres", "All Steam genres")
	t.addAssetMark()
	t.Genres = genres
	t.Date = config.Value
	t.Prices = prices

	returnTemplate(w, r, "stats_genres", t)
}

type statsGenresTemplate struct {
	globalTemplate
	Genres []mysql.Genre
	Date   string
	Prices map[int]string
}
