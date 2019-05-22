package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
)

func APIRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", apiHandler)
	r.Get("/app/{id}", apiApp)
	r.Get("/package/{id}", apiPackage)
	r.Get("/bundle/{id}", apiBundle)
	r.Get("/group/{id}", apiGroup)
	r.Get("/player/{id}", apiPlayer)
	return r
}

func apiHandler(w http.ResponseWriter, r *http.Request) {

	t := apiTemplate{}
	t.fill(w, r, "API", "")

	err := returnTemplate(w, r, "api", t)
	log.Err(err, r)
}

type apiTemplate struct {
	GlobalTemplate
	Calls []apiCall
}

type apiCall struct {
	path   string
	params []apiCallParam
}

type apiCallParam struct {
	name string
	typ  string
}

func apiApp(w http.ResponseWriter, r *http.Request) {

}

func apiPackage(w http.ResponseWriter, r *http.Request) {

}

func apiBundle(w http.ResponseWriter, r *http.Request) {

}

func apiPlayer(w http.ResponseWriter, r *http.Request) {

}

func apiGroup(w http.ResponseWriter, r *http.Request) {

}
