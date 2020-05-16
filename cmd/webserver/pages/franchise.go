package pages

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
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

	publisher, err := sql.GetPublisher(id)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Invalid App ID"})
		return
	}

	fmt.Println(publisher.GetName())
}
