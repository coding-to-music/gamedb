package tasks

import (
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
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

func (c AppsQueueInflux) Group() TaskGroup {
	return TaskGroupApps
}

func (c AppsQueueInflux) Cron() TaskTime {
	return CronTimeAppsInflux
}

func (c AppsQueueInflux) work() (err error) {

	var projection = bson.M{"_id": 1}

	return mongo.BatchApps(nil, projection, func(apps []mongo.App) {

		var ids []int
		for _, v := range apps {
			ids = append(ids, v.ID)
		}

		var chunks = helpers.ChunkInts(ids, 20)

		for _, chunk := range chunks {
			err = queue.ProduceAppsInflux(chunk)
			if err != nil {
				log.Err(err)
				return
			}
		}
	})
}
