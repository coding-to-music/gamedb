package pages

import (
	"net/http"

	"github.com/gamedb/website/pkg/log"
	"github.com/go-chi/chi"
)

func esiRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/header", headerHandler)
	return r
}

func headerHandler(w http.ResponseWriter, r *http.Request) {

	setCacheHeaders(w, 0)

	t := GlobalTemplate{}
	t.fill(w, r, "Header", "")

	err := returnTemplate(w, r, "_header_esi", t)
	log.Err(err, r)
}
