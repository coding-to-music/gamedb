package tasks

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsYoutube struct {
	BaseTask
}

func (c AppsYoutube) ID() string {
	return "apps-youtube"
}

func (c AppsYoutube) Name() string {
	return "Queue top apps for youtube stats"
}

func (c AppsYoutube) Cron() string {
	return CronTimeAppsYoutube
}

func (c AppsYoutube) work() (err error) {

	apps, err := mongo.GetApps(0, 100, bson.D{{"player_peak_week", -1}}, nil, bson.M{"_id": 1, "name": 1}, nil)
	if err != nil {
		return err
	}

	for _, app := range apps {

		err = queue.ProduceAppsYoutube(app.ID, app.Name)
		log.Err(err)
	}

	return nil
}
