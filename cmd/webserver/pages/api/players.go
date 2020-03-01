package api

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/webserver/pages/api/generated"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func (s Server) GetPlayers(w http.ResponseWriter, r *http.Request) {

	s.call(w, r, func(w http.ResponseWriter, r *http.Request) (code int, response interface{}) {

		params := generated.ParamsForGetPlayers(r.Context())

		var limit int64 = 10
		if params.Limit != nil {
			limit = int64(*params.Limit)
		}

		var offset int64 = 0
		if params.Offset != nil {
			offset = int64(*params.Offset)
		}

		filter := bson.D{{}}

		if params.Continent != nil {
			filter = append(filter, bson.E{Key: "continent_code", Value: *params.Continent})
		}

		if params.Country != nil {
			filter = append(filter, bson.E{Key: "country_code", Value: *params.Country})
		}

		players, err := mongo.GetPlayers(offset, limit, nil, filter, nil)
		if err != nil {
			return 500, err
		}

		total, err := mongo.CountDocuments(mongo.CollectionPlayers, filter, 0)
		if err != nil {
			log.Err(err, r)
		}

		result := generated.PlayersResponse{}
		result.Pagination.Fill(offset, limit, total)

		for _, player := range players {

			result.Players = append(result.Players, generated.PlayerSchema{
				Id:   player.ID,
				Name: player.GetName(),
			})
		}

		return 200, result
	})
}
