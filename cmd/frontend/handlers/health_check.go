package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func HealthCheckRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", healthCheckHandler)

	return r
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
}
