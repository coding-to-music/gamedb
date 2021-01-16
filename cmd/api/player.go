package main

import (
	"errors"
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
		returnErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	player, err := mongo.GetPlayer(id)
	if err == mongo.ErrNoDocuments {

		ua := r.UserAgent()
		err = queue.ProducePlayer(queue.PlayerMessage{ID: id, UserAgent: &ua}, "api-retrieve")
		if err != nil {
			log.ErrS(err)
		}

		returnErrorResponse(w, http.StatusNotFound, errors.New("player not found, queued"))
		return

	} else if err != nil {

		log.ErrS(err)
		returnErrorResponse(w, http.StatusInternalServerError, err)
		return

	} else {

		ret := generated.PlayerResponse{}
		ret.Id = strconv.FormatInt(player.ID, 10)
		ret.Name = player.GetName()
		ret.Avatar = player.GetAvatar()

		ret.Continent = player.ContinentCode
		ret.Country = player.CountryCode
		ret.State = player.StateCode

		ret.Badges = player.BadgesCount
		ret.Comments = player.CommentsCount
		ret.Friends = player.FriendsCount
		ret.Games = player.GamesCount
		ret.Level = player.Level
		ret.Playtime = player.PlayTime
		ret.Groups = player.GroupsCount

		returnResponse(w, http.StatusOK, ret)
	}
}
