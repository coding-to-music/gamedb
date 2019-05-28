package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
)

func BadgeRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", badgeHandler)
	r.Get("/{slug}", badgeHandler)
	return r
}

func badgeHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid group ID"})
		return
	}

	t := badgeTemplate{}
	t.fill(w, r, "Badge", "")

	err := returnTemplate(w, r, "badge", t)
	log.Err(err, r)
}

type badgeTemplate struct {
	GlobalTemplate
}
