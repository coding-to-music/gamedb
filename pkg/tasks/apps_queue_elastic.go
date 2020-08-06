package tasks

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsQueueElastic struct {
	BaseTask
}

func (c AppsQueueElastic) ID() string {
	return "apps-queue-elastic"
}

func (c AppsQueueElastic) Name() string {
	return "Queue all apps to Elastic"
}

func (c AppsQueueElastic) Group() TaskGroup {
	return TaskGroupElastic
}

func (c AppsQueueElastic) Cron() TaskTime {
	return ""
}

func (c AppsQueueElastic) work() (err error) {

	var projection = bson.M{
		"common":       0,
		"config":       0,
		"extended":     0,
		"install":      0,
		"launch":       0,
		"localization": 0,
		"reviews":      0,
		"ufs":          0,
	}

	return mongo.BatchApps(nil, projection, func(apps []mongo.App) {

		for _, app := range apps {

			err = queue.ProduceAppSearch(&app, 0)
			if err != nil {
				log.Err(err)
				return
			}
		}
	})
}
