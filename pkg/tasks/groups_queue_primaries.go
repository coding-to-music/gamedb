package tasks

import (
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type PlayersQueuePrimaries struct {
	BaseTask
}

func (c PlayersQueuePrimaries) ID() string {
	return "groups-queue-primaries"
}

func (c PlayersQueuePrimaries) Name() string {
	return "Queue all group primaries"
}

func (c PlayersQueuePrimaries) Group() string {
	return TaskGroupGroups
}

func (c PlayersQueuePrimaries) Cron() string {
	return ""
}

func (c PlayersQueuePrimaries) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		groups, err := mongo.GetGroups(limit, offset, bson.D{{"_id", 1}}, nil, bson.M{"_id": 1})
		if err != nil {
			return err
		}

		for _, group := range groups {

			err = queue.ProduceGroupPrimaries(group.ID)
			if err != nil {
				return err
			}
		}

		if int64(len(groups)) != limit {
			break
		}

		offset += limit
	}

	return nil
}
