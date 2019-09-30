package pages

import (
	"net/http"

	"github.com/go-chi/chi"
)

func InfoRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", infoHandler)
	return r
}

func infoHandler(w http.ResponseWriter, r *http.Request) {

	t := GlobalTemplate{}
	t.fill(w, r, "Info", "")

	returnTemplate(w, r, "info", t)
}
