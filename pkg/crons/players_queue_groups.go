package crons

import (
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type PlayersQueueGroups struct {
	BaseTask
}

func (c PlayersQueueGroups) ID() string {
	return "players-queue-groups"
}

func (c PlayersQueueGroups) Name() string {
	return "Refresh player-groups for all players"
}

func (c PlayersQueueGroups) Group() TaskGroup {
	return TaskGroupPlayers
}

func (c PlayersQueueGroups) Cron() TaskTime {
	return ""
}

func (c PlayersQueueGroups) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		var projection = bson.M{
			"common":       0,
			"config":       0,
			"extended":     0,
			"install":      0,
			"launch":       0,
			"localization": 0,
			"reviews":      0,
			"ufs":          0,
		}

		players, err := mongo.GetPlayers(offset, limit, nil, nil, projection)
		if err != nil {
			return err
		}

		for _, player := range players {

			err = queue.ProducePlayerGroup(player, true, false)
			if err != nil {
				return err
			}
		}

		if int64(len(players)) != limit {
			break
		}

		offset += limit
	}

	return nil
}
