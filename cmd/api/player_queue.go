package main

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/queue"
)

func (s Server) PostPlayersId(w http.ResponseWriter, r *http.Request, id int64) {

	s.call(w, r, func(w http.ResponseWriter, r *http.Request) (code int, response interface{}) {

		err := queue.ProducePlayer(queue.PlayerMessage{ID: id})
		if err == memcache.ErrInQueue {
			return 200, err
		} else if err != nil {
			log.ErrS(err)
			return 500, err
		}

		return 200, "Player queued"
	})
}
