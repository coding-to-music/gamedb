package web

import (
	"net/http"
	"time"

	"github.com/gamedb/website/log"
)

func apiHandler(w http.ResponseWriter, r *http.Request) {

	setCacheHeaders(w, time.Hour*24*7)

	t := apiTemplate{}
	t.fill(w, r, "API", "")

	err := returnTemplate(w, r, "api", t)
	log.Err(err, r)
}

type apiTemplate struct {
	GlobalTemplate
	Commits []commit
	Hash    string
}
