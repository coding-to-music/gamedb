package web

import (
	"net/http"

	"github.com/gamedb/website/log"
)

func donateHandler(w http.ResponseWriter, r *http.Request) {

	t := GlobalTemplate{}
	t.Fill(w, r, "Donate", "Help pay for the server costs or just buy me a beer.")

	err := returnTemplate(w, r, "donate", t)
	log.Err(err)
}
