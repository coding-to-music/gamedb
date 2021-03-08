package crons

import (
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/crons/helpers/rabbitweb"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsSameOwners struct {
	BaseTask
}

func (c AppsSameOwners) ID() string {
	return "apps-sameowners"
}

func (c AppsSameOwners) Name() string {
	return "Queue a game to scan same owners"
}

func (c AppsSameOwners) Group() TaskGroup {
	return TaskGroupApps
}

func (c AppsSameOwners) Cron() TaskTime {
	return CronTimeAppsSameowners
}

func (c AppsSameOwners) work() (err error) {

	queues, err := rabbitweb.GetRabbitWebClient().GetQueues()
	if err != nil {
		return err
	}

	var free int
	for _, v := range queues {
		if v.Name == string(consumers.QueueAppsSameowners) {
			free = int(float64(v.Consumers)/float64(consumers.ConsumersPerProcess)) - v.Messages
			break
		}
	}

	if free <= 0 {
		return nil
	}

	filter := bson.D{
		{"group_followers", bson.M{"$gte": 1000}},
		{"owners", bson.M{"$gte": 0}},
	}

	sort := bson.D{
		{"related_owners_app_ids_date", 1},
		{"group_followers", 1},
	}

	apps, err := mongo.GetApps(0, int64(free), sort, filter, bson.M{"_id": 1})
	for _, v := range apps {

		err = consumers.ProduceSameOwners(v.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
