package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
)

func BadgesRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", badgesHandler)
	r.Mount("/{id}", BadgeRouter())
	return r
}

func badgesHandler(w http.ResponseWriter, r *http.Request) {

	t := badgesTemplate{}
	t.fill(w, r, "Badges", "Steam badge leaderboards")

	var err error
	t.Badges, err = mongo.GetBadgeSummaries()
	log.Err(err, r)

	returnTemplate(w, r, "badges", t)
}

type badgesTemplate struct {
	GlobalTemplate
	Badges []mongo.PlayerBadgeSummary
}
