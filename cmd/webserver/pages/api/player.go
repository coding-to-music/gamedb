package api

import (
	"errors"
	"net/http"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
)

func (s Server) PostPlayersId(w http.ResponseWriter, r *http.Request) {

	s.call(w, r, func(w http.ResponseWriter, r *http.Request) (code int, response interface{}) {

		if id, ok := r.Context().Value("id").(int64); ok {

			err := queue.ProducePlayer(queue.PlayerMessage{ID: id, SkipGroups: true})

			// todo, handle different errors properly
			if err != nil {
				log.Err(err)
				return 500, err
			}

			return 200, "Player queued"
		}

		return 400, errors.New("invalid app ID")
	})
}
