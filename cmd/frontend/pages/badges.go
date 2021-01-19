package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
)

func BadgesRouter() http.Handler {

	r := chi.NewRouter()
	r.Mount("/{id}", BadgeRouter())

	r.Get("/", badgesHandler)

	return r
}

func badgesHandler(w http.ResponseWriter, r *http.Request) {

	t := badgesTemplate{}
	t.fill(w, r, "badges", "Steam badge leaderboards", "See who's the highst badge level, and who got it first")

	var err error
	t.Badges, err = mongo.GetBadgeSummaries()
	if err != nil {
		log.ErrS(err)
	}

	returnTemplate(w, r, t)
}

type badgesTemplate struct {
	globalTemplate
	Badges []mongo.PlayerBadgeSummary
}
