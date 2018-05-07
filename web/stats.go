package web

import (
	"net/http"
)

func StatsHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t:= statsTemplate{}
	t.Fill(w, r, "Stats")

	returnTemplate(w, r, "stats", t)
}

type statsTemplate struct {
	GlobalTemplate
}
