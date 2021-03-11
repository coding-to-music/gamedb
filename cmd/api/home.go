package main

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/pkg/config"
)

func (s Server) Get(w http.ResponseWriter, r *http.Request) {

	returnResponse(w, r, http.StatusOK, generated.HomeResponse{
		Docs: config.C.GlobalSteamDomain + "/api/globalsteam",
	})
}
