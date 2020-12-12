package main

import (
	"net/http"
	"strconv"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func (s Server) GetGroups(w http.ResponseWriter, r *http.Request, params generated.GetGroupsParams) {

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
			log.ErrS(err)
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
