package tasks

import (
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

func (c AppsQueueElastic) Group() string {
	return TaskGroupElastic
}

func (c AppsQueueElastic) Cron() string {
	return ""
}

func (c AppsQueueElastic) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

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

		apps, err := mongo.GetApps(offset, limit, bson.D{{"_id", 1}}, nil, projection)
		if err != nil {
			return err
		}

		for _, app := range apps {

			err = queue.ProduceAppSearch(app)
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
