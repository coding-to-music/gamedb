package api

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/webserver/pages/api/generated"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func (s Server) GetApps(w http.ResponseWriter, r *http.Request) {

	s.call(w, r, func(w http.ResponseWriter, r *http.Request) (code int, response interface{}) {

		params := generated.ParamsForGetApps(r.Context())

		var limit int64 = 10
		if params.Limit != nil {
			limit = int64(*params.Limit)
		}

		var offset int64 = 0
		if params.Offset != nil {
			offset = int64(*params.Offset)
		}

		filter := bson.D{{}}

		if params.Ids != nil {
			filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$in": *params.Ids}})
		}

		if params.Tags != nil {
			filter = append(filter, bson.E{Key: "tags", Value: bson.M{"$in": *params.Tags}})
		}

		if params.Genres != nil {
			filter = append(filter, bson.E{Key: "genres", Value: bson.M{"$in": *params.Genres}})
		}

		if params.Categories != nil {
			filter = append(filter, bson.E{Key: "categories", Value: bson.M{"$in": *params.Categories}})
		}

		if params.Developers != nil {
			filter = append(filter, bson.E{Key: "developers", Value: bson.M{"$in": *params.Developers}})
		}

		if params.Publishers != nil {
			filter = append(filter, bson.E{Key: "publishers", Value: bson.M{"$in": *params.Publishers}})
		}

		if params.Platforms != nil {
			filter = append(filter, bson.E{Key: "platforms", Value: bson.M{"$in": *params.Platforms}})
		}

		apps, err := mongo.GetApps(offset, limit, nil, filter, nil, nil)
		if err != nil {
			return 500, err
		}

		total, err := mongo.CountDocuments(mongo.CollectionApps, filter, 0)
		if err != nil {
			log.Err(err, r)
		}

		result := generated.AppsResponse{}
		result.Pagination.Fill(offset, limit, total)

		for _, app := range apps {

			result.Apps = append(result.Apps, generated.AppSchema{
				Id:   app.ID,
				Name: app.GetName(),
			})
		}

		return 200, result
	})
}
