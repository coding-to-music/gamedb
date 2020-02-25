package api

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/webserver/pages/api/generated"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
)

func (s Server) PostPlayersId(w http.ResponseWriter, r *http.Request) {

	if id, ok := r.Context().Value("id").(int64); ok {

		err := queue.ProducePlayer(queue.PlayerMessage{ID: id, SkipGroups: true})
		// todo, handle different errors properly
		if err != nil {
			log.Err(err)
			s.ReturnError(w, 500, err.Error())
			return
		}

		payload := generated.SucccessSchema{Message: "Player queued"}
		s.Return200(w, payload)
	}
}
