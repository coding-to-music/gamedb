package main

import (
	"errors"
	"net/http"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/pkg/backend"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.uber.org/zap"
)

func (s Server) GetGamesId(w http.ResponseWriter, r *http.Request, id int32, params generated.GetGamesIdParams) {

	s.call(w, r, func(w http.ResponseWriter, r *http.Request) (code int, response interface{}) {

		if id, ok := r.Context().Value("id").(int32); ok {

			app, err := mongo.GetApp(int(id))
			if err == mongo.ErrNoDocuments {

				return 404, errors.New("app not found")

			} else if err != nil {

				zap.S().Error(err)
				return 500, err

			} else {

				ret := generated.AppResponse{}
				ret.Id = app.ID
				ret.Name = app.GetName()
				ret.ReleaseDate = app.ReleaseDateUnix

				ret.Genres = app.Genres
				ret.Tags = app.Tags
				ret.Categories = app.Categories
				ret.Publishers = app.Publishers
				ret.Developers = app.Developers

				ret.PlayersMax = app.PlayerPeakAllTime
				ret.PlayersWeekMax = app.PlayerPeakWeek
				ret.PlayersWeekAvg = app.PlayerAverageWeek

				ret.ReviewsNegative = app.Reviews.Positive
				ret.ReviewsPositive = app.Reviews.Negative
				ret.ReviewsScore = app.ReviewsScore
				ret.MetacriticScore = int32(app.MetacriticScore)

				for _, v := range app.Prices {
					ret.Prices = append(ret.Prices, struct {
						Currency        string `json:"currency"`
						DiscountPercent int32  `json:"discountPercent"`
						Final           int32  `json:"final"`
						Free            bool   `json:"free"`
						Individual      int32  `json:"individual"`
						Initial         int32  `json:"initial"`
					}{
						Currency:        string(v.Currency),
						DiscountPercent: int32(v.DiscountPercent),
						Final:           int32(v.Final),
						Free:            v.Free,
						Individual:      int32(v.Individual),
						Initial:         int32(v.Initial),
					})
				}

				return 200, ret
			}
		}

		return 400, errors.New("invalid app ID")
	})
}

func (s Server) GetGames(w http.ResponseWriter, r *http.Request, params generated.GetGamesParams) {

	s.call(w, r, func(w http.ResponseWriter, r *http.Request) (code int, response interface{}) {

		var limit int64 = 10
		if params.Limit != nil && *params.Limit >= 1 && *params.Limit <= 1000 {
			limit = int64(*params.Limit)
		}

		var offset int64 = 0
		if params.Offset != nil {
			offset = int64(*params.Offset)
		}

		payload := &backend.ListAppsRequest{}
		payload.Offset = offset
		payload.Limit = limit

		if params.Ids != nil {
			// payload.Ids = *params.Ids
		}

		if params.Tags != nil {
			// payload.Tags = *params.Ids
		}

		if params.Genres != nil {
			// payload.Genres = *params.Genres
		}

		if params.Categories != nil {
			// payload.Categories = *params.Categories
		}

		if params.Developers != nil {
			// payload.Developers = *params.Developers
		}

		if params.Publishers != nil {
			// payload.Publishers = *params.Publishers
		}

		if params.Platforms != nil {
			payload.Platforms = *params.Platforms
		}
		//
		// var projection = bson.M{
		// 	"id":                  1,
		// 	"name":                1,
		// 	"tags":                1,
		// 	"genres":              1,
		// 	"developers":          1,
		// 	"categories":          1,
		// 	"prices":              1,
		// 	"player_peak_alltime": 1,
		// 	"player_peak_week":    1,
		// 	"player_avg_week":     1,
		// 	"release_date_unix":   1,
		// 	"reviews":             1,
		// 	"reviews_score":       1,
		// }
		//
		// apps, err := mongo.GetApps(offset, limit, bson.D{{"_id", 1}}, filter, projection)
		// if err != nil {
		// 	return 500, err
		// }
		//
		// total, err := mongo.CountDocuments(mongo.CollectionApps, filter, 0)
		// if err != nil {
		// 	zap.S().Error(err)
		// }
		//
		// result := generated.AppsResponse{}
		// result.Pagination.Fill(offset, limit, total)
		//
		// for _, app := range apps {
		//
		// 	result.Apps = append(result.Apps, generated.AppSchema{
		// 		Id:   app.ID,
		// 		Name: app.GetName(),
		// 	})
		// }

		conn, ctx, err := backend.GetClient()
		if err != nil {
			return 500, err
		}

		resp, err := backend.NewAppsServiceClient(conn).Apps(ctx, payload)
		if err != nil {
			return 500, err
		}

		result := generated.AppsResponse{}
		result.Pagination.Fill(offset, limit, resp.Pagination.GetTotal())

		for _, app := range resp.Apps {

			result.Apps = append(result.Apps, generated.AppSchema{
				Id:   int(app.GetId()),
				Name: app.GetName(),
			})
		}

		return 200, result
	})
}
