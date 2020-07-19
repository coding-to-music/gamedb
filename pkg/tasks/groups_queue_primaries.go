package tasks

import (
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type GroupsQueuePrimaries struct {
	BaseTask
}

func (c GroupsQueuePrimaries) ID() string {
	return "groups-queue-primaries"
}

func (c GroupsQueuePrimaries) Name() string {
	return "Queue all group primaries to be updated"
}

func (c GroupsQueuePrimaries) Group() string {
	return TaskGroupGroups
}

func (c GroupsQueuePrimaries) Cron() string {
	return ""
}

func (c GroupsQueuePrimaries) work() (err error) {

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
