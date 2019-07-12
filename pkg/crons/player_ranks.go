package crons

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
)

type PlayerRanks struct {
}

func (c PlayerRanks) ID() CronEnum {
	return CronPlayerRanks
}

func (c PlayerRanks) Name() string {
	return "Update player ranks"
}

func (c PlayerRanks) Config() sql.ConfigType {
	return sql.ConfRanksUpdated
}

func (c PlayerRanks) Work() {

	started(c)

	cronLogInfo("Level")
	err := mongo.RankPlayers("level", "level_rank")
	log.Warning(err)

	cronLogInfo("Games")
	err = mongo.RankPlayers("games_count", "games_rank")
	log.Warning(err)

	cronLogInfo("Badges")
	err = mongo.RankPlayers("badges_count", "badges_rank")
	log.Warning(err)

	cronLogInfo("Time")
	err = mongo.RankPlayers("play_time", "play_time_rank")
	log.Warning(err)

	cronLogInfo("Friends")
	err = mongo.RankPlayers("friends_count", "friends_rank")
	log.Warning(err)

	//
	finished(c)
}
