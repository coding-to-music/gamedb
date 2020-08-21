package main

import (
	"errors"
	"net/http"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.uber.org/zap"
)

func (s Server) PostPlayersId(w http.ResponseWriter, r *http.Request, id int64, params generated.PostPlayersIdParams) {

	s.call(w, r, func(w http.ResponseWriter, r *http.Request) (code int, response interface{}) {

		if id, ok := r.Context().Value("id").(int64); ok {

			err := queue.ProducePlayer(queue.PlayerMessage{ID: id})
			if err == memcache.ErrInQueue {
				return 200, err
			} else if err != nil {
				zap.S().Error(err)
				return 500, err
			}

			return 200, "Player queued"
		}

		return 400, errors.New("invalid app ID")
	})
}
