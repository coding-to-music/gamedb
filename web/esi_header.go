package web

import "net/http"

func HeaderHandler(w http.ResponseWriter, r *http.Request) {

	t := GlobalTemplate{}
	t.Fill(w, r, "Header")

	returnTemplate(w, r, "_header_esi", t)
}
