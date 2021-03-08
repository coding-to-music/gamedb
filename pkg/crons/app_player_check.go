package crons

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/crons/helpers/rabbitweb"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsPlayerCheck struct {
	BaseTask
}

func (c AppsPlayerCheck) ID() string {
	return "app-players"
}

func (c AppsPlayerCheck) Name() string {
	return "Check apps for players (Bottom)"
}

func (c AppsPlayerCheck) Group() TaskGroup {
	return TaskGroupApps
}

func (c AppsPlayerCheck) Cron() TaskTime {
	return CronTimeAppPlayers
}

func (c AppsPlayerCheck) work() (err error) {

	// Skip if queues have activity
	limits := map[rabbit.QueueName]int{
		consumers.QueueAppPlayers: 1000,
	}

	queues, err := rabbitweb.GetRabbitWebClient().GetQueues()
	if err != nil {
		return err
	}

	for _, q := range queues {
		if val, ok := limits[rabbit.QueueName(q.Name)]; ok && q.Messages > val {
			return nil
		}
	}

	var filter = bson.D{
		{"player_peak_week", bson.M{"$lt": topAppPlayers}},
		{"group_followers", bson.M{"$lt": topGroupFollowers}},
	}

	var projection = bson.M{"_id": 1}

	return mongo.BatchApps(filter, projection, func(apps []mongo.App) {

		var ids []int
		for _, v := range apps {
			if v.ID > 0 { // This is just here to stop storing things on app 0, which we use to store steam stats on
				ids = append(ids, v.ID)
			}
		}

		var chunks = helpers.ChunkInts(ids, 20)

		for _, chunk := range chunks {
			err = consumers.ProduceAppPlayers(chunk)
			if err != nil {
				log.ErrS(err)
				return
			}
		}
	})
}
