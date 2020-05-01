package tasks

import (
	"strconv"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AppPlayers struct {
	BaseTask
}

func (c AppPlayers) ID() string {
	return "app-players"
}

func (c AppPlayers) Name() string {
	return "Check apps for players"
}

func (c AppPlayers) Cron() string {
	return CronTimeAppPlayers
}

func (c AppPlayers) work() (err error) {

	// Check queue size
	q, err := queue.Channels[rabbit.Producer][queue.QueueAppPlayers].Inspect()
	if err != nil {
		return err
	}

	if q.Messages > 100_000 {
		return nil
	}

	// Add apps to queue
	var offset int64 = 0
	var limit int64 = 10_000

	for {

		apps, err := mongo.GetApps(offset, limit, bson.D{{"_id", 1}}, nil, bson.M{"_id": 1}, nil)
		if err != nil {
			return err
		}

		log.Info("Found " + strconv.Itoa(len(apps)) + " apps")

		var ids []int
		for _, v := range apps {
			if v.ID > 0 { // This is just here to stop storing things on app 0, which we use to store steam stats on
				ids = append(ids, v.ID)
			}
		}

		var chunks = helpers.ChunkInts(ids, 50)

		for _, chunk := range chunks {
			err = queue.ProduceAppPlayers(queue.AppPlayerMessage{IDs: chunk})
			log.Err(err)
		}

		if int64(len(apps)) != limit {
			break
		}

		offset += limit
	}

	return nil
}
