package main

import (
	"errors"
	"net/http"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
)

func (s Server) GetGamesId(w http.ResponseWriter, r *http.Request, id int32) {

	s.call(w, r, func(w http.ResponseWriter, r *http.Request) (code int, response interface{}) {

		app, err := mongo.GetApp(int(id))
		if err == mongo.ErrNoDocuments {

			return 404, errors.New("app not found")

		} else if err != nil {

			log.ErrS(err)
			return 500, err

		} else {

			ret := generated.AppResponse{}
			ret.Id = app.ID
			ret.Name = app.GetName()
			ret.ReleaseDate = app.ReleaseDateUnix
			ret.PlayersMax = app.PlayerPeakAllTime
			ret.PlayersWeekMax = app.PlayerPeakWeek
			ret.ReviewsNegative = app.Reviews.Positive
			ret.ReviewsPositive = app.Reviews.Negative
			ret.ReviewsScore = app.ReviewsScore
			ret.MetacriticScore = int32(app.MetacriticScore)
			// ret.PlayersWeekAvg = app.PlayerAverageWeek

			// Fix nulls in JSON
			ret.Prices = generated.AppSchema_Prices{
				AdditionalProperties: map[string]generated.ProductPriceSchema{},
			}

			for k, v := range app.Prices {
				ret.Prices.AdditionalProperties[string(k)] = generated.ProductPriceSchema{
					Currency:        string(v.Currency),
					DiscountPercent: int32(v.DiscountPercent),
					Final:           int32(v.Final),
					Free:            v.Free,
					Individual:      int32(v.Individual),
					Initial:         int32(v.Initial),
				}
			}

			categories, err := app.GetCategories()
			if err != nil {
				log.ErrS(err)
			}
			for _, v := range categories {
				ret.Categories = append(ret.Categories, generated.StatSchema{Id: v.ID, Name: v.Name})
			}

			tags, err := app.GetTags()
			if err != nil {
				log.ErrS(err)
			}
			for _, v := range tags {
				ret.Tags = append(ret.Tags, generated.StatSchema{Id: v.ID, Name: v.Name})
			}

			genres, err := app.GetGenres()
			if err != nil {
				log.ErrS(err)
			}
			for _, v := range genres {
				ret.Genres = append(ret.Genres, generated.StatSchema{Id: v.ID, Name: v.Name})
			}

			publishers, err := app.GetPublishers()
			if err != nil {
				log.ErrS(err)
			}
			for _, v := range publishers {
				ret.Publishers = append(ret.Publishers, generated.StatSchema{Id: v.ID, Name: v.Name})
			}

			developers, err := app.GetDevelopers()
			if err != nil {
				log.ErrS(err)
			}
			for _, v := range developers {
				ret.Developers = append(ret.Developers, generated.StatSchema{Id: v.ID, Name: v.Name})
			}

			return 200, ret
		}
	})
}
