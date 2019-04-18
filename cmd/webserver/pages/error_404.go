package pages

import (
	"net/http"
)

func Error404Handler(w http.ResponseWriter, r *http.Request) {

	setCacheHeaders(w, 0)

	returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "This page doesnt exist"})
}
