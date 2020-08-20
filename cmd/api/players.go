package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
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

func (s Server) GetPlayers(w http.ResponseWriter, r *http.Request, params generated.GetPlayersParams) {

	s.call(w, r, func(w http.ResponseWriter, r *http.Request) (code int, response interface{}) {

		var limit int64 = 10
		if params.Limit != nil && *params.Limit >= 1 && *params.Limit <= 1000 {
			limit = int64(*params.Limit)
		}

		var offset int64 = 0
		if params.Offset != nil {
			offset = int64(*params.Offset)
		}

		var sort string
		if params.Sort != nil {
			switch *params.Sort {
			case "id":
				sort = "_id"
			case "level":
				sort = "level"
			case "badges":
				sort = "badges_count"
			case "games":
				sort = "games_count"
			case "time":
				sort = "play_time"
			case "friends":
				sort = "friends_count"
			case "comments":
				sort = "comments_count"
			default:
				sort = "_id"
			}
		}

		var order int
		if params.Order != nil {
			switch *params.Sort {
			case "1", "asc", "ascending":
				order = 1
			case "0", "-1", "desc", "descending":
				order = -1
			default:
				order = -1
			}
		}

		filter := bson.D{{}}

		if params.Continent != nil {
			filter = append(filter, bson.E{Key: "continent_code", Value: *params.Continent})
		}

		if params.Country != nil {
			filter = append(filter, bson.E{Key: "country_code", Value: *params.Country})
		}

		players, err := mongo.GetPlayers(offset, limit, bson.D{{sort, order}}, filter, bson.M{"_id": 1,
			"persona_name":   1,
			"avatar":         1,
			"continent_code": 1,
			"country_code":   1,
			"status_code":    1,
			"badges_count":   1,
			"comments_count": 1,
			"friends_count":  1,
			"games_count":    1,
			"groups_count":   1,
			"level":          1,
			"play_time":      1})
		if err != nil {
			return 500, err
		}

		total, err := mongo.CountDocuments(mongo.CollectionPlayers, filter, 0)
		if err != nil {
			zap.S().Error(err)
		}

		result := generated.PlayersResponse{}
		result.Pagination.Fill(offset, limit, total)

		for _, player := range players {

			result.Players = append(result.Players, generated.PlayerSchema{
				Id:     strconv.FormatInt(player.ID, 10),
				Name:   player.PersonaName,
				Avatar: player.Avatar,

				Continent: player.ContinentCode,
				Country:   player.CountryCode,
				State:     player.StateCode,

				Badges:   player.BadgesCount,
				Comments: player.CommentsCount,
				Friends:  player.FriendsCount,
				Games:    player.GamesCount,
				Level:    player.Level,
				Playtime: player.PlayTime,
				Groups:   player.GroupsCount,
			})
		}

		return 200, result
	})
}
