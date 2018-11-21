package web

import (
	"net/http"

	"github.com/gamedb/website/logging"
)

func donateHandler(w http.ResponseWriter, r *http.Request) {

	t := GlobalTemplate{}
	t.Fill(w, r, "Donate")
	t.Description = "Help pay for the server costs or just buy me a beer."

	err := returnTemplate(w, r, "donate", t)
	logging.Error(err)
}
