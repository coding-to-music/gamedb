package web

import (
	"net/http"
	"time"

	"github.com/gamedb/website/log"
)

func infoHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*24)

	t := GlobalTemplate{}
	t.fill(w, r, "Info", "")

	err := returnTemplate(w, r, "info", t)
	log.Err(err, r)
}
