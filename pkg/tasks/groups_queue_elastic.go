package tasks

import (
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type GroupsQueueElastic struct {
	BaseTask
}

func (c GroupsQueueElastic) ID() string {
	return "groups-queue-elastic"
}

func (c GroupsQueueElastic) Name() string {
	return "Queue all groups to Elastic"
}

func (c GroupsQueueElastic) Group() string {
	return TaskGroupGroups
}

func (c GroupsQueueElastic) Cron() string {
	return ""
}

func (c GroupsQueueElastic) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		var projection = bson.M{}
		var filter = bson.D{{"type", helpers.GroupTypeGroup}}

		groups, err := mongo.GetGroups(limit, offset, bson.D{{"_id", 1}}, filter, projection)
		if err != nil {
			return err
		}

		for _, group := range groups {

			err = queue.ProduceGroupSearch(group)
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
