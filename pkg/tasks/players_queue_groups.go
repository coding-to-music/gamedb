package tasks

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

func (c PlayersQueueGroups) Group() string {
	return TaskGroupPlayers
}

func (c PlayersQueueGroups) Cron() string {
	return ""
}

func (c PlayersQueueGroups) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		var projection = bson.M{
			"_id":          1,
			"persona_name": 1,
			"avatar":       1,
		}

		players, err := mongo.GetPlayers(offset, limit, bson.D{{"_id", 1}}, nil, projection)
		if err != nil {
			return err
		}

		for _, player := range players {

			err = queue.ProducePlayerGroup(player.ID, player.PersonaName, player.Avatar, true)
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
