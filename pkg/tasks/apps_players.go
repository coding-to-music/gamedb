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

	if q.Messages > 1000 {
		return nil
	}

	// Add apps to queue
	apps, err := mongo.GetApps(0, 0, bson.D{{"_id", 1}}, nil, bson.M{"_id": 1}, nil)
	if err != nil {
		return err
	}

	var ids []int
	for _, v := range apps {
		if v.ID > 0 { // This is just here to stop storing things on app 0, which we use to store steam stats on
			ids = append(ids, v.ID)
		}
	}

	log.Info("Found " + strconv.Itoa(len(ids)) + " apps")

	idChunks := helpers.ChunkInts(ids, 10)

	for _, idChunk := range idChunks {

		err = queue.ProduceAppPlayers(queue.AppPlayerMessage{IDs: idChunk})
		log.Err(err)
	}

	return nil
}
