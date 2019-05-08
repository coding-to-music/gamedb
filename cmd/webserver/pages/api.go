package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
)

func APIRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", apiHandler)
	return r
}

func apiHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	t := apiTemplate{}
	t.fill(w, r, "API", "")

	err := returnTemplate(w, r, "api", t)
	log.Err(err, r)
}

type apiTemplate struct {
	GlobalTemplate
	Commits []commit
	Hash    string
}
