package web

import (
	"net/http"

	"github.com/gamedb/website/log"
	"github.com/go-chi/chi"
)

func depotsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", depotsHandler)
	r.Get("/{id}", depotHandler)
	return r
}

func depotsHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := depotsTemplate{}
	t.Fill(w, r, "Depots", "")

	err := returnTemplate(w, r, "depots", t)
	log.Log(err)
}

type depotsTemplate struct {
	GlobalTemplate
}

func depotHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := depotTemplate{}
	t.Fill(w, r, "Depot", "")

	err := returnTemplate(w, r, "depot", t)
	log.Log(err)
}

type depotTemplate struct {
	GlobalTemplate
}
