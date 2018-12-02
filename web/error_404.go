package web

import "net/http"

func error404Handler(w http.ResponseWriter, r *http.Request) {
	returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "This page doesnt exist"})
}
