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
	return "apps-reindex-elastic"
}

func (c AppsQueueElastic) Name() string {
	return "Reindex all apps in Elastic"
}

func (c AppsQueueElastic) Cron() string {
	return ""
}

func (c AppsQueueElastic) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		var projection = bson.M{
			"common":       -1,
			"config":       -1,
			"extended":     -1,
			"install":      -1,
			"launch":       -1,
			"localization": -1,
			"reviews":      -1,
			"ufs":          -1,
		}

		apps, err := mongo.GetApps(offset, limit, bson.D{{"_id", 1}}, nil, projection, nil)
		if err != nil {
			return err
		}

		for _, app := range apps {

			err = queue.ProduceAppSearch(app)
			log.Err(err)
		}

		if int64(len(apps)) != limit {
			break
		}

		offset += limit
	}

	return nil
}
