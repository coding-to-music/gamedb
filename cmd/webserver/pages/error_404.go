package pages

import (
	"net/http"
)

func Error404Handler(w http.ResponseWriter, r *http.Request) {

	returnErrorTemplate(w, r, errorTemplate{Code: http.StatusNotFound, Message: "This page doesnt exist"})
}
