package pages

import (
	"net/http"

	"github.com/go-chi/chi"
)

func TermsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", termsHandler)
	return r
}

func termsHandler(w http.ResponseWriter, r *http.Request) {

	t := globalTemplate{}
	t.fill(w, r, "terms", "Terms", "Terms of Service")

	returnTemplate(w, r, t)
}
