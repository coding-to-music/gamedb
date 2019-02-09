package web

import (
	"net/http"

	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
)

func steamAPIHandler(w http.ResponseWriter, r *http.Request) {

	resp, _, err := helpers.GetSteam().GetSupportedAPIList()
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "Can't talk to Steam"})
		return
	}

	t := steamAPITemplate{}
	t.Fill(w, r, "Steam API", "")
	t.Interfaces = resp.Interfaces

	err = returnTemplate(w, r, "steam_api", t)
	log.Err(err, r)
}

type steamAPITemplate struct {
	GlobalTemplate
	Interfaces interface{}
}
