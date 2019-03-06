package web

import (
	"net/http"

	"github.com/gamedb/website/log"
)

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("OK"))
	log.Err(err, r)
}
