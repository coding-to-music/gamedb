package tasks

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsUpdateYoutube struct {
	BaseTask
}

func (c AppsUpdateYoutube) ID() string {
	return "apps-update-youtube"
}

func (c AppsUpdateYoutube) Name() string {
	return "Queue top apps for youtube stats"
}

func (c AppsUpdateYoutube) Cron() string {
	return CronTimeAppsYoutube
}

func (c AppsUpdateYoutube) work() (err error) {

	apps, err := mongo.GetApps(0, 90, bson.D{{"player_peak_week", -1}}, nil, bson.M{"_id": 1, "name": 1}, nil)
	if err != nil {
		return err
	}

	for _, app := range apps {

		err = queue.ProduceAppsYoutube(app.ID, app.Name)
		log.Err(err)
	}

	return nil
}
