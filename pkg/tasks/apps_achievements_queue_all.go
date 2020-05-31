package tasks

import (
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsAchievementsQueueAll struct {
	BaseTask
}

func (c AppsAchievementsQueueAll) ID() string {
	return "queue-all-app-achievements"
}

func (c AppsAchievementsQueueAll) Name() string {
	return "Queue all app achievements"
}

func (c AppsAchievementsQueueAll) Cron() string {
	return ""
}

func (c AppsAchievementsQueueAll) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		apps, err := mongo.GetApps(offset, limit, bson.D{{"_id", 1}}, nil, bson.M{"_id": 1})
		if err != nil {
			return err
		}

		for _, app := range apps {

			err = queue.ProduceAppAchievement(app.ID)
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
