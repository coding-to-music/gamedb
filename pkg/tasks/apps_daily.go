package tasks

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsDaily struct {
	BaseTask
}

func (c AppsDaily) ID() string {
	return "apps-daily"
}

func (c AppsDaily) Name() string {
	return "Queue all apps daily"
}

func (c AppsDaily) Cron() string {
	return CronTimeAppsDaily
}

func (c AppsDaily) work() (err error) {

	// apps, err := mongo.GetApps(0, 0, bson.D{{"_id", 1}}, nil, bson.M{"_id": 1, "name": 1}, nil)
	apps, err := mongo.GetApps(0, 10, bson.D{{"player_peak_week", -1}}, nil, bson.M{"_id": 1, "name": 1}, nil)
	if err != nil {
		return err
	}

	for _, app := range apps {

		err = queue.ProduceAppsDaily(app.ID, app.Name)
		log.Err(err)
	}

	return nil
}
