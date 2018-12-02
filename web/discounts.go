package web

import (
	"net/http"

	"github.com/gamedb/website/log"
)

func discountsHandler(w http.ResponseWriter, r *http.Request) {

	t := discountsTemplate{}
	t.Fill(w, r, "Discounts", "")

	err := returnTemplate(w, r, "discounts", t)
	log.Log(err)
}

type discountsTemplate struct {
	GlobalTemplate
}
