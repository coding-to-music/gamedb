package tasks

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/rabbitweb"
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

func (c AppsPlayerCheck) Group() string {
	return TaskGroupApps
}

func (c AppsPlayerCheck) Cron() string {
	return CronTimeAppPlayers
}

func (c AppsPlayerCheck) work() (err error) {

	// Skip if queues have activity
	limits := map[rabbit.QueueName]int{
		queue.QueueAppPlayers: 1000,
	}

	queues, err := rabbitweb.RabbitClient.GetQueues()
	if err != nil {
		return err
	}

	for _, q := range queues {
		if val, ok := limits[rabbit.QueueName(q.Name)]; ok && q.Messages > val {
			return nil
		}
	}

	// Add apps to queue
	var offset int64 = 0
	var limit int64 = 10_000
	var filter = bson.D{{"player_peak_week", bson.M{"$lt": topAppPlayers}}}

	for {

		apps, err := mongo.GetApps(offset, limit, bson.D{{"_id", 1}}, filter, bson.M{"_id": 1})
		if err != nil {
			return err
		}

		var ids []int
		for _, v := range apps {
			if v.ID > 0 { // This is just here to stop storing things on app 0, which we use to store steam stats on
				ids = append(ids, v.ID)
			}
		}

		var chunks = helpers.ChunkInts(ids, 20)

		for _, chunk := range chunks {
			err = queue.ProduceAppPlayers(chunk)
			if err != nil {
				return err
			}
		}

		if int64(len(apps)) != limit {
			break
		}

		offset += limit
	}

	return nil
}
