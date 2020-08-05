package tasks

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsQueueWishlists struct {
	BaseTask
}

func (c AppsQueueWishlists) ID() string {
	return "apps-queue-wishlists"
}

func (c AppsQueueWishlists) Name() string {
	return "Update wishlist stats for all apps"
}

func (c AppsQueueWishlists) Group() TaskGroup {
	return TaskGroupApps
}

func (c AppsQueueWishlists) Cron() TaskTime {
	return CronTimeAppsWishlists
}

func (c AppsQueueWishlists) work() (err error) {

	var projection = bson.M{"_id": 1}

	return mongo.BatchApps(nil, projection, func(apps []mongo.App) {

		for _, app := range apps {

			err = queue.ProduceAppsWishlists(app.ID)
			if err != nil {
				log.Err(err)
				return
			}
		}
	})
}
