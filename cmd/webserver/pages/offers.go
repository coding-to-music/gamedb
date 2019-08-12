package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
)

func OffersRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", offersHandler)
	return r
}

func offersHandler(w http.ResponseWriter, r *http.Request) {

	t := GlobalTemplate{}
	t.fill(w, r, "Offers", "")

	err := returnTemplate(w, r, "offers", t)
	log.Err(err, r)
}
