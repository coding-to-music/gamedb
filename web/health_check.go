package web

import (
	"net/http"

	"github.com/gamedb/website/log"
)

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("OK"))
	log.Log(err)
}
