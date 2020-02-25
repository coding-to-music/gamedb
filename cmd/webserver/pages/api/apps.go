package api

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/webserver/pages/api/generated"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func (s Server) GetApps(w http.ResponseWriter, r *http.Request) {

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

	}

	if params.Tags != nil {

	}

	apps, err := mongo.GetApps(offset, limit, nil, filter, nil, nil)
	if err != nil {
		s.ReturnError(w, 500, err.Error())
		return
	}

	result := generated.Apps{}
	result.Pagination.Fill(r.Context())

	for _, app := range apps {

		result.Apps = append(result.Apps, generated.AppSchema{
			Id:   app.ID,
			Name: app.GetName(),
		})
	}

	s.Return200(w, result)
}
