package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
)

func HealthCheckRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", healthCheckHandler)
	return r
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("OK"))
	log.Err(err, r)
}
