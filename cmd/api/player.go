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

	s.call(w, r, func(w http.ResponseWriter, r *http.Request) (code int, response interface{}) {

		id, err := helpers.IsValidPlayerID(id)
		if err != nil {
			return 404, err
		}

		player, err := mongo.GetPlayer(id)
		if err == mongo.ErrNoDocuments {

			ua := r.UserAgent()
			err = queue.ProducePlayer(queue.PlayerMessage{ID: id, UserAgent: &ua}, "api-retrieve")
			if err != nil {
				log.ErrS(err)
			}

			return 404, "player not found, trying to add player"

		} else if err != nil {

			log.ErrS(err)
			return 500, err

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

			return 200, ret
		}
	})
}
