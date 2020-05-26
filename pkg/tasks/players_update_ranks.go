package tasks

import (
	"github.com/gamedb/gamedb/pkg/helpers/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
)

type PlayersUpdateRanks struct {
	BaseTask
}

func (c PlayersUpdateRanks) ID() string {
	return "update-player-ranks"
}

func (c PlayersUpdateRanks) Name() string {
	return "Update player ranks"
}

func (c PlayersUpdateRanks) Cron() string {
	return CronTimePlayerRanks
}

func (c PlayersUpdateRanks) work() (err error) {

	// Global
	for read, write := range mongo.PlayerRankFields {
		err = queue.ProducePlayerRank(queue.PlayerRanksMessage{
			SortColumn: read,
			ObjectKey:  string(write),
		})
		if err != nil {
			return err
		}
	}

	// Continents
	for _, continent := range i18n.Continents {
		for read, write := range mongo.PlayerRankFields {
			err = queue.ProducePlayerRank(queue.PlayerRanksMessage{
				SortColumn: read,
				ObjectKey:  string(write) + "_continent-" + continent.Key,
				Continent:  &continent.Key,
			})
			if err != nil {
				return err
			}
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
			if err != nil {
				return err
			}
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
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
