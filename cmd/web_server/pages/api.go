package pages

import (
	"net/http"
	"time"

	"github.com/gamedb/website/pkg/log"
	"github.com/go-chi/chi"
)

func apiRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", apiHandler)
	return r
}

func apiHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*24)

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
