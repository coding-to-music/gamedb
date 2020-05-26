package tasks

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsAchievementsQueueElastic struct {
	BaseTask
}

func (c AppsAchievementsQueueElastic) ID() string {
	return "achievements-queue-elastic"
}

func (c AppsAchievementsQueueElastic) Name() string {
	return "Queue all achievements to Elastic"
}

func (c AppsAchievementsQueueElastic) Cron() string {
	return ""
}

func (c AppsAchievementsQueueElastic) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	var apps = map[int]mongo.App{} // Memory cache

	for {

		appAchievements, err := mongo.GetAppAchievements(offset, limit, nil, bson.D{{"_id", 1}})
		if err != nil {
			return err
		}

		for _, appAchievement := range appAchievements {

			var app mongo.App
			if val, ok := apps[appAchievement.AppID]; ok {
				app = val
			} else {
				app, err = mongo.GetApp(appAchievement.AppID)
				if err != nil {
					log.Err(err)
					continue
				}
				apps[appAchievement.AppID] = app
			}

			err = queue.ProduceAchievementSearch(appAchievement, app)
			if err != nil {
				return err
			}
		}

		if int64(len(appAchievements)) != limit {
			break
		}

		offset += limit
	}

	return nil
}
