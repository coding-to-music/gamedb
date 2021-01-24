package main

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
)

func (s Server) GetGamesId(w http.ResponseWriter, r *http.Request, id int32) {

	app, err := mongo.GetApp(int(id))
	if err == mongo.ErrNoDocuments {

		returnResponse(w, r, http.StatusNotFound, generated.GameResponse{Error: "app not found"})
		return

	} else if err != nil {

		log.ErrS(err)
		returnResponse(w, r, http.StatusInternalServerError, generated.GameResponse{Error: err.Error()})
		return

	} else {

		gameSchema := generated.GameSchema{}
		gameSchema.Id = app.ID
		gameSchema.Name = app.Name
		gameSchema.Icon = app.Icon
		gameSchema.ReleaseDate = app.ReleaseDateUnix
		gameSchema.PlayersMax = app.PlayerPeakAllTime
		gameSchema.PlayersWeekMax = app.PlayerPeakWeek
		gameSchema.ReviewsNegative = app.Reviews.Positive
		gameSchema.ReviewsPositive = app.Reviews.Negative
		gameSchema.ReviewsScore = app.ReviewsScore
		gameSchema.MetacriticScore = int32(app.MetacriticScore)
		// ret.PlayersWeekAvg = app.PlayerAverageWeek

		// Fix nulls in JSON
		gameSchema.Prices = generated.GameSchema_Prices{
			AdditionalProperties: map[string]generated.ProductPriceSchema{},
		}

		for k, v := range app.Prices {
			gameSchema.Prices.AdditionalProperties[string(k)] = generated.ProductPriceSchema{
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
			gameSchema.Categories = append(gameSchema.Categories, generated.StatSchema{Id: v.ID, Name: v.Name})
		}

		tags, err := app.GetTags()
		if err != nil {
			log.ErrS(err)
		}
		for _, v := range tags {
			gameSchema.Tags = append(gameSchema.Tags, generated.StatSchema{Id: v.ID, Name: v.Name})
		}

		genres, err := app.GetGenres()
		if err != nil {
			log.ErrS(err)
		}
		for _, v := range genres {
			gameSchema.Genres = append(gameSchema.Genres, generated.StatSchema{Id: v.ID, Name: v.Name})
		}

		publishers, err := app.GetPublishers()
		if err != nil {
			log.ErrS(err)
		}
		for _, v := range publishers {
			gameSchema.Publishers = append(gameSchema.Publishers, generated.StatSchema{Id: v.ID, Name: v.Name})
		}

		developers, err := app.GetDevelopers()
		if err != nil {
			log.ErrS(err)
		}
		for _, v := range developers {
			gameSchema.Developers = append(gameSchema.Developers, generated.StatSchema{Id: v.ID, Name: v.Name})
		}

		returnResponse(w, r, http.StatusOK, generated.GameResponse{Game: gameSchema})
	}
}
