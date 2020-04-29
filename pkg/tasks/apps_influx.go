package tasks

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsInflux struct {
	BaseTask
}

func (c AppsInflux) ID() string {
	return "apps-influx"
}

func (c AppsInflux) Name() string {
	return "Update app peaks and averages"
}

func (c AppsInflux) Cron() string {
	return CronTimeAppsInflux
}

func (c AppsInflux) work() (err error) {

	apps, err := mongo.GetApps(0, 0, bson.D{{"_id", 1}}, nil, bson.M{"_id": 1}, nil)
	if err != nil {
		return err
	}

	for _, app := range apps {

		err = queue.ProduceAppsInflux(app.ID)
		log.Err(err)
	}

	return nil
}
