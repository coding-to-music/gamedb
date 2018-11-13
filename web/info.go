package web

import "net/http"

func InfoHandler(w http.ResponseWriter, r *http.Request) {

	t := GlobalTemplate{}
	t.Fill(w, r, "Info")

	returnTemplate(w, r, "info", t)
}
