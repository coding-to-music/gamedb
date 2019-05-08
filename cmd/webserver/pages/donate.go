package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
)

func DonateRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", donateHandler)
	return r
}

func donateHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	t := GlobalTemplate{}
	t.fill(w, r, "Donate", "Databases take up a tonne of memory and space. Help pay for the server costs or just buy me a beer.")

	err := returnTemplate(w, r, "donate", t)
	log.Err(err, r)
}
