package tasks

import (
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type PlayerRanks struct {
	BaseTask
}

func (c PlayerRanks) ID() string {
	return "update-player-ranks"
}

func (c PlayerRanks) Name() string {
	return "Update player ranks"
}

func (c PlayerRanks) Cron() string {
	return CronTimePlayerRanks
}

func (c PlayerRanks) work() (err error) {

	// Fix nulls
	_, err = mongo.UpdateManySet(mongo.CollectionPlayers, bson.D{{"ranks", nil}}, bson.D{{"ranks", bson.M{}}})
	if err != nil {
		return err
	}

	var fields = map[string]mongo.RankMetric{
		"level":          mongo.RankKeyLevel,
		"games_count":    mongo.RankKeyGames,
		"badges_count":   mongo.RankKeyBadges,
		"play_time":      mongo.RankKeyPlaytime,
		"friends_count":  mongo.RankKeyFriends,
		"comments_count": mongo.RankKeyComments,
	}

	// Global
	for read, write := range fields {

		msg := consumers.PlayerRanksMessage{}
		msg.SortColumn = read
		msg.ObjectKey = string(write)

		err = msg.Produce()
		log.Err(err)
	}

	// Continents
	for _, continent := range helpers.Continents {

		for read, write := range fields {

			msg := consumers.PlayerRanksMessage{}
			msg.SortColumn = read
			msg.ObjectKey = string(write) + "_continent-" + continent.Key
			msg.Continent = &continent.Key

			err = msg.Produce()
			log.Err(err)
		}
	}

	// Countries
	countryCodes, err := mongo.GetUniquePlayerCountries()
	if err != nil {
		return err
	}

	for _, cc := range countryCodes {

		for read, write := range fields {

			msg := consumers.PlayerRanksMessage{}
			msg.SortColumn = read
			msg.ObjectKey = string(write) + "_country-" + cc
			msg.Country = &cc

			err = msg.Produce()
			log.Err(err)
		}
	}

	// Rank by State
	for _, cc := range mongo.CountriesWithStates {

		stateCodes, err := mongo.GetUniquePlayerStates(cc)
		if err != nil {
			log.Err(err)
			continue
		}

		for _, state := range stateCodes {

			for read, write := range fields {

				msg := consumers.PlayerRanksMessage{}
				msg.SortColumn = read
				msg.ObjectKey = string(write) + "_state-" + state.Key
				msg.Country = &cc
				msg.State = &state.Key

				err = msg.Produce()
				log.Err(err)
			}
		}
	}

	return nil
}
