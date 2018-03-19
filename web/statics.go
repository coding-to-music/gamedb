package web

import "net/http"

func InfoHandler(w http.ResponseWriter, r *http.Request) {

	template := staticTemplate{}
	template.Fill(r, "Info")

	returnTemplate(w, r, "info", template)
}

func DonateHandler(w http.ResponseWriter, r *http.Request) {

	template := staticTemplate{}
	template.Fill(r, "Donate")

	returnTemplate(w, r, "donate", template)
}

func Error404Handler(w http.ResponseWriter, r *http.Request) {

	returnErrorTemplate(w, r, 404, "Page not found")
}

type staticTemplate struct {
	GlobalTemplate
}
