package web

import (
	"net/http"
)

func DiscountsHandler(w http.ResponseWriter, r *http.Request) {

	template := discountsTemplate{}
	template.Fill(r, "Discounts")

	returnTemplate(w, r, "discounts", template)
	return
}

type discountsTemplate struct {
	GlobalTemplate
}
