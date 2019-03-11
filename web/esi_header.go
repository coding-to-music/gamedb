package web

import (
	"net/http"

	"github.com/gamedb/website/log"
)

func headerHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	t := GlobalTemplate{}
	t.fill(w, r, "Header", "")

	err := returnTemplate(w, r, "_header_esi", t)
	log.Err(err, r)
}
