package web

import (
	"net/http"

	"github.com/gamedb/website/logging"
)

func infoHandler(w http.ResponseWriter, r *http.Request) {

	t := GlobalTemplate{}
	t.Fill(w, r, "Info")

	err:=returnTemplate(w, r, "info", t)
	logging.Error(err)
}
