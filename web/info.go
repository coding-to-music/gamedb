package web

import (
	"net/http"
	"time"

	"github.com/gamedb/website/log"
)

func infoHandler(w http.ResponseWriter, r *http.Request) {

	setCacheHeaders(w, time.Hour*24*7)

	t := GlobalTemplate{}
	t.fill(w, r, "Info", "")

	err := returnTemplate(w, r, "info", t)
	log.Err(err, r)
}
