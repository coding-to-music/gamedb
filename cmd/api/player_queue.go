package main

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/queue"
)

func (s Server) PostPlayersId(w http.ResponseWriter, r *http.Request, id int64) {

	err := queue.ProducePlayer(queue.PlayerMessage{ID: id}, "api-update")
	if err == memcache.ErrInQueue {

		returnResponse(w, r, http.StatusOK, generated.MessageResponse{Message: "Already in queue"})
		return

	} else if err != nil {

		log.ErrS(err)
		returnResponse(w, r, http.StatusInternalServerError, generated.MessageResponse{Error: err.Error()})
		return
	}

	returnResponse(w, r, http.StatusOK, generated.MessageResponse{Message: "Player queued"})
}
