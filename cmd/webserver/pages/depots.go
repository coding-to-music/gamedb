package pages

import (
	"net/http"
	"strconv"

	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi"
)

func DepotsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", depotsHandler)
	r.Get("/{id}", depotHandler)
	return r
}

func depotsHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := depotsTemplate{}
	t.fill(w, r, "Depots", "Steam depots")

	returnTemplate(w, r, "depots", t)
}

type depotsTemplate struct {
	GlobalTemplate
}

func depotHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid Depot ID"})
		return
	}

	// Template
	t := depotTemplate{}
	t.fill(w, r, "Depot", "Steam depot")
	t.Depot = mysql.Depot{}
	t.Depot.ID = id

	returnTemplate(w, r, "depot", t)
}

type depotTemplate struct {
	GlobalTemplate
	Depot mysql.Depot
}
