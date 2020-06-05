package tasks

import (
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsPlayerCheck struct {
	BaseTask
}

func (c AppsPlayerCheck) ID() string {
	return "app-players"
}

func (c AppsPlayerCheck) Name() string {
	return "Check apps for players"
}

func (c AppsPlayerCheck) Cron() string {
	return CronTimeAppPlayers
}

func (c AppsPlayerCheck) work() (err error) {

	// Check queue size
	q, err := queue.ProducerChannels[queue.QueueAppPlayers].Inspect()
	if err != nil {
		return err
	}

	if q.Messages > 500 {
		return nil
	}

	// Add apps to queue
	var offset int64 = 0
	var limit int64 = 10_000

	for {

		apps, err := mongo.GetApps(offset, limit, bson.D{{"_id", 1}}, nil, bson.M{"_id": 1})
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
			err = queue.ProduceAppPlayers(queue.AppPlayerMessage{IDs: chunk})
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
