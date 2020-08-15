package pages

import (
	"net/http"

	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func HealthCheckRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", healthCheckHandler)
	return r
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("OK"))
	zap.S().Error(err)
}
