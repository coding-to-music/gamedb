package web

import (
	"net/http"
	"strconv"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
	"github.com/go-chi/chi"
)

func depotsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", depotsHandler)
	r.Get("/{id}", depotHandler)
	return r
}

func depotHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid Depot ID."})
		return
	}

	idx, err := strconv.Atoi(id)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid Depot ID: " + id})
		return
	}

	// Template
	t := depotTemplate{}
	t.Fill(w, r, "Depot", "")
	t.Depot = db.Depot{}
	t.Depot.ID = idx

	err = returnTemplate(w, r, "depot", t)
	log.Log(err)
}

type depotTemplate struct {
	GlobalTemplate
	Depot db.Depot
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
