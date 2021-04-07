package crons

import (
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
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

func (c PlayersUpdateRanks) Group() TaskGroup {
	return TaskGroupPlayers
}

func (c PlayersUpdateRanks) Cron() TaskTime {
	return CronTimePlayerRanks
}

func (c PlayersUpdateRanks) work() (err error) {

	// Global
	for read, write := range helpers.PlayerRankFields {
		err = consumers.ProducePlayerRank(consumers.PlayerRanksMessage{
			SortColumn: read,
			ObjectKey:  string(write),
		})
		if err != nil {
			return err
		}
	}

	// By continent
	for _, continent := range i18n.Continents {
		for read, write := range helpers.PlayerRankFields {
			err = consumers.ProducePlayerRank(consumers.PlayerRanksMessage{
				SortColumn: read,
				ObjectKey:  string(write) + "_continent-" + continent.Key,
				Continent:  &continent.Key,
			})
			if err != nil {
				return err
			}
		}
	}

	// By country
	for cc, states := range i18n.States {

		for read, write := range helpers.PlayerRankFields {
			err = consumers.ProducePlayerRank(consumers.PlayerRanksMessage{
				SortColumn: read,
				ObjectKey:  string(write) + "_country-" + cc,
				Country:    &cc,
			})
			if err != nil {
				return err
			}
		}

		// By state
		for state := range states {
			for read, write := range helpers.PlayerRankFields {
				err = consumers.ProducePlayerRank(consumers.PlayerRanksMessage{
					SortColumn: read,
					ObjectKey:  string(write) + "_state-" + state,
					Country:    &cc,
					State:      &state,
				})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
