package tasks

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
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
	return "2 0"
}

func (c PlayerRanks) work() {

	log.Info("Level")
	err := mongo.RankPlayers("level", "level_rank")
	log.Warning(err)

	log.Info("Games")
	err = mongo.RankPlayers("games_count", "games_rank")
	log.Warning(err)

	log.Info("Badges")
	err = mongo.RankPlayers("badges_count", "badges_rank")
	log.Warning(err)

	log.Info("Time")
	err = mongo.RankPlayers("play_time", "play_time_rank")
	log.Warning(err)

	log.Info("Friends")
	err = mongo.RankPlayers("friends_count", "friends_rank")
	log.Warning(err)
}
