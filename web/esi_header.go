package web

import (
	"net/http"

	"github.com/gamedb/website/logging"
)

func headerHandler(w http.ResponseWriter, r *http.Request) {

	t := GlobalTemplate{}
	t.Fill(w, r, "Header")

	err := returnTemplate(w, r, "_header_esi", t)
	logging.Error(err)
}
