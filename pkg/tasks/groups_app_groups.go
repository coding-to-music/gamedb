package tasks

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type QueueAppGroups struct {
	BaseTask
}

func (c QueueAppGroups) ID() string {
	return "queue-app-groups"
}

func (c QueueAppGroups) Name() string {
	return "Queue app groups"
}

func (c QueueAppGroups) Cron() string {
	return CronTimeQueueAppGroups
}

func (c QueueAppGroups) work() (err error) {

	apps, err := mongo.GetApps(0, 0, nil, bson.D{{"group_id", bson.M{"$ne": ""}}}, bson.M{"group_id": 1}, nil)
	if err != nil {
		return err
	}

	for _, app := range apps {

		err = queue.ProduceGroup(queue.GroupMessage{ID: app.GroupID})
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		if err != nil {
			log.Err(err)
		}
	}

	//
	log.Info(strconv.Itoa(len(apps)) + " groups queued")

	apps = nil // Memory

	return nil
}
