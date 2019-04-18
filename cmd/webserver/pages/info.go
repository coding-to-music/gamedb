package pages

import (
	"net/http"
	"time"

	"github.com/gamedb/website/pkg/log"
	"github.com/go-chi/chi"
)

func InfoRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", infoHandler)
	return r
}

func infoHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*24)

	t := GlobalTemplate{}
	t.fill(w, r, "Info", "")

	err := returnTemplate(w, r, "info", t)
	log.Err(err, r)
}
