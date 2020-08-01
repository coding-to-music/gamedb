package api

import (
	"errors"
	"net/http"

	"github.com/gamedb/gamedb/cmd/webserver/pages/api/generated"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
)

func (s Server) GetGamesId(w http.ResponseWriter, r *http.Request) {

	s.call(w, r, func(w http.ResponseWriter, r *http.Request) (code int, response interface{}) {

		if id, ok := r.Context().Value("id").(int32); ok {

			app, err := mongo.GetApp(int(id))
			if err == mongo.ErrNoDocuments {

				return 404, errors.New("app not found")

			} else if err != nil {

				log.Err(err, r)
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
