package tasks

import (
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsQueueInflux struct {
	BaseTask
}

func (c AppsQueueInflux) ID() string {
	return "apps-queue-influx"
}

func (c AppsQueueInflux) Name() string {
	return "Update app peaks and averages (influx)"
}

func (c AppsQueueInflux) Group() string {
	return TaskGroupApps
}

func (c AppsQueueInflux) Cron() string {
	return CronTimeAppsInflux
}

func (c AppsQueueInflux) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		apps, err := mongo.GetApps(offset, limit, bson.D{{"_id", 1}}, nil, bson.M{"_id": 1})
		if err != nil {
			return err
		}

		var ids []int
		for _, v := range apps {
			ids = append(ids, v.ID)
		}

		var chunks = helpers.ChunkInts(ids, 20)

		for _, chunk := range chunks {
			err = queue.ProduceAppsInflux(chunk)
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
