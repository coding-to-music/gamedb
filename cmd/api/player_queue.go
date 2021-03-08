package main

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/log"
)

func (s Server) PostPlayersId(w http.ResponseWriter, r *http.Request, id int64) {

	err := consumers.ProducePlayer(consumers.PlayerMessage{ID: id}, "api-update")
	if err == consumers.ErrInQueue {

		returnResponse(w, r, http.StatusOK, generated.MessageResponse{Message: "Already in queue"})
		return

	} else if err != nil {

		log.ErrS(err)
		returnResponse(w, r, http.StatusInternalServerError, generated.MessageResponse{Error: err.Error()})
		return
	}

	returnResponse(w, r, http.StatusOK, generated.MessageResponse{Message: "Player queued"})
}
