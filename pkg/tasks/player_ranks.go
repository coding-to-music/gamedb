package tasks

import (
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
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

	// Global
	for read, write := range mongo.PlayerRankFields {
		err = queue.ProducePlayerRank(queue.PlayerRanksMessage{
			SortColumn: read,
			ObjectKey:  string(write),
		})
		log.Err(err)
	}

	// Continents
	for _, continent := range helpers.Continents {
		for read, write := range mongo.PlayerRankFields {
			err = queue.ProducePlayerRank(queue.PlayerRanksMessage{
				SortColumn: read,
				ObjectKey:  string(write) + "_continent-" + continent.Key,
				Continent:  &continent.Key,
			})
			log.Err(err)
		}
	}

	// Countries
	countryCodes, err := mongo.GetUniquePlayerCountries()
	if err != nil {
		return err
	}

	for _, cc := range countryCodes {
		for read, write := range mongo.PlayerRankFields {
			err = queue.ProducePlayerRank(queue.PlayerRanksMessage{
				SortColumn: read,
				ObjectKey:  string(write) + "_country-" + cc,
				Country:    &cc,
			})
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
			for read, write := range mongo.PlayerRankFields {
				err = queue.ProducePlayerRank(queue.PlayerRanksMessage{
					SortColumn: read,
					ObjectKey:  string(write) + "_state-" + state.Key,
					Country:    &cc,
					State:      &state.Key,
				})
				log.Err(err)
			}
		}
	}

	return nil
}
