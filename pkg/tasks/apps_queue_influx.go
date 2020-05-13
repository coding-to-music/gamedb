package tasks

import (
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

func (c AppsQueueInflux) Cron() string {
	return CronTimeAppsInflux
}

func (c AppsQueueInflux) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		apps, err := mongo.GetApps(offset, limit, bson.D{{"_id", 1}}, nil, bson.M{"_id": 1}, nil)
		if err != nil {
			return err
		}

		// log.Info(strconv.Itoa(len(apps)) + " apps")

		for _, app := range apps {

			err = queue.ProduceAppsInflux(app.ID)
			log.Err(err)
		}

		if int64(len(apps)) != limit {
			break
		}

		offset += limit
	}

	return nil
}
