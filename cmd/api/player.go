package main

import (
	"net/http"
	"strconv"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
)

func (s Server) GetPlayersId(w http.ResponseWriter, r *http.Request, id int64) {

	id, err := helpers.IsValidPlayerID(id)
	if err != nil {
		returnResponse(w, r, http.StatusBadRequest, generated.PlayerResponse{Error: err.Error()})
		return
	}

	player, err := mongo.GetPlayer(id)
	if err == mongo.ErrNoDocuments {

		ua := r.UserAgent()
		err = queue.ProducePlayer(queue.PlayerMessage{ID: id, UserAgent: &ua}, "api-retrieve")
		if err != nil {
			log.ErrS(err)
		}

		returnResponse(w, r, http.StatusNotFound, generated.PlayerResponse{Error: "player not found, queued"})
		return

	} else if err != nil {

		log.ErrS(err)
		returnResponse(w, r, http.StatusInternalServerError, generated.PlayerResponse{Error: err.Error()})
		return

	} else {

		playerSchema := generated.PlayerSchema{}
		playerSchema.Id = strconv.FormatInt(player.ID, 10)
		playerSchema.Name = player.GetName()
		playerSchema.Avatar = player.GetAvatar()

		playerSchema.Continent = player.ContinentCode
		playerSchema.Country = player.CountryCode
		playerSchema.State = player.StateCode

		playerSchema.Badges = player.BadgesCount
		playerSchema.Comments = player.CommentsCount
		playerSchema.Friends = player.FriendsCount
		playerSchema.Games = player.GamesCount
		playerSchema.Level = player.Level
		playerSchema.Playtime = player.PlayTime
		playerSchema.Groups = player.GroupsCount

		returnResponse(w, r, http.StatusOK, generated.PlayerResponse{Player: playerSchema})
	}
}
