package main

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/pkg/helpers"
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

		var sort = "members"
		if params.Sort != nil {
			switch *params.Sort {
			case "id":
				sort = "_id"
			case "members":
				sort = "members"
			case "trending":
				sort = "trending"
			case "primaries":
				sort = "primaries"
			default:
				sort = "members"
			}
		}

		var order = -1
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

		filter := bson.D{{Key: "type", Value: helpers.GroupTypeGroup}}

		if params.Ids != nil {
			filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$in": *params.Ids}})
		}

		projection := bson.M{
			"_id":             1,
			"name":            1,
			"abbreviation":    1,
			"url":             1,
			"app_id":          1,
			"headline":        1,
			"icon":            1,
			"trending":        1,
			"members":         1,
			"members_in_chat": 1,
			"members_in_game": 1,
			"members_online":  1,
			"error":           1,
			"type":            1,
			"primaries":       1,
		}

		groups, err := mongo.GetGroups(offset, limit, bson.D{{sort, order}}, filter, projection)
		if err != nil {
			return 500, err
		}

		total, err := mongo.CountDocuments(mongo.CollectionGroups, filter, 0)
		if err != nil {
			log.ErrS(err)
		}

		result := generated.GroupsResponse{}
		result.Pagination.Fill(offset, limit, total)

		for _, group := range groups {

			result.Groups = append(result.Groups, generated.GroupSchema{
				Abbreviation:  group.GetAbbr(),
				AppId:         int32(group.AppID),
				Error:         group.Error,
				Headline:      group.Headline,
				Icon:          group.GetIcon(),
				Id:            group.ID,
				Members:       int32(group.Members),
				MembersInChat: int32(group.MembersInChat),
				MembersInGame: int32(group.MembersInGame),
				MembersOnline: int32(group.MembersOnline),
				Name:          group.GetName(),
				Primaries:     int32(group.Primaries),
				Trending:      float32(group.Trending),
				Url:           group.GetURL(),
			})
		}

		return 200, result
	})
}
