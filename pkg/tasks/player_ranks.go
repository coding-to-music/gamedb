package tasks

import (
	"time"

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
	return CronTimePlayerRanks
}

func (c PlayerRanks) work() {

	ranks := []rankTask{
		{"Level", "level", "level_rank"},
		{"Games", "games_count", "games_rank"},
		{"Badges", "badges_count", "badges_rank"},
		{"Time", "play_time", "play_time_rank"},
		{"Friends", "friends_count", "friends_rank"},
	}

	for _, v := range ranks {

		log.Info(v.name)

		err := mongo.RankPlayers(v.readCol, v.writeCol)
		log.Warning(err)

		time.Sleep(time.Second * 30)
	}
}

type rankTask struct {
	name     string
	readCol  string
	writeCol string
}

func (rt rankTask) getWriteCol(cc string) string {
	return rt.writeCol + cc
}
