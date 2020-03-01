package api

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/log"
)

func (s Server) GetPlayers(w http.ResponseWriter, r *http.Request) {

	s.call(w, r, func(w http.ResponseWriter, r *http.Request) (code int, response interface{}) {

		log.Info("players coming soon")

		return 200, "players"
	})
}
