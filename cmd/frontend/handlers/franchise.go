package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi/v5"
)

func FranchiseRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", franchisesHandler)
	r.Get("/{id}", franchiseHandler)
	return r
}

func franchisesHandler(w http.ResponseWriter, r *http.Request) {

}

func franchiseHandler(w http.ResponseWriter, r *http.Request) {

	// Get publisher
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid App ID"})
		return
	}

	publisher, err := mongo.GetStat(mongo.StatsTypePublishers, id)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Invalid App ID"})
		return
	}

	fmt.Println(publisher.Name)
}
