package web

import (
	"net/http"

	"github.com/gamedb/website/logging"
)

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("OK"))
	logging.Error(err)
}
