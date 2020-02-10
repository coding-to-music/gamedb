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

type QueuePlayerGroups struct {
	BaseTask
}

func (c QueuePlayerGroups) ID() string {
	return "queue-player-groups"
}

func (c QueuePlayerGroups) Name() string {
	return "Queue player groups"
}

func (c QueuePlayerGroups) Cron() string {
	return CronTimeQueuePlayerGroups
}

func (c QueuePlayerGroups) work() (err error) {

	var filter = bson.D{
		{Key: "type", Value: helpers.GroupTypeGroup},
	}

	var sorts = []bson.D{
		{{"members", -1}},
		{{"trending", 1}},
		{{"trending", -1}},
	}

	var groupMap = map[string]bool{}

	for _, sort := range sorts {

		groups, err := mongo.GetGroups(1000, 0, sort, filter, bson.M{"_id": 1})
		if err != nil {
			return err
		}

		for _, group := range groups {
			groupMap[group.ID] = true
		}
	}

	for groupID := range groupMap {

		err = queue.ProduceGroup(queue.GroupMessage{ID: groupID})
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		if err != nil {
			log.Err(err)
		}
	}

	//
	log.Info(strconv.Itoa(len(groupMap)) + " groups queued")

	return nil
}
