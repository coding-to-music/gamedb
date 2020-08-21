package main

import (
	"net/http"
	"strconv"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.uber.org/zap"
)

func (s Server) GetPlayersId(w http.ResponseWriter, r *http.Request, id int64, params generated.GetPlayersIdParams) {

	s.call(w, r, func(w http.ResponseWriter, r *http.Request) (code int, response interface{}) {

		if id, ok := r.Context().Value("id").(int64); ok {

			id, err := helpers.IsValidPlayerID(id)
			if err != nil {
				return 404, err
			}

			player, err := mongo.GetPlayer(id)
			if err == mongo.ErrNoDocuments {

				ua := r.UserAgent()
				err = queue.ProducePlayer(queue.PlayerMessage{ID: id, UserAgent: &ua})
				if err != nil {
					zap.S().Error(err)
				}

				return 404, "player not found, trying to add player"

			} else if err != nil {

				zap.S().Error(err)
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
		}

		return 400, "invalid player ID"
	})
}
