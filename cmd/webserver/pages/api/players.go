package api

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/log"
)

func (s Server) GetPlayers(w http.ResponseWriter, r *http.Request) {
	log.Info("players")
}
