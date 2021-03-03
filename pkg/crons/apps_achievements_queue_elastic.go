package crons

import (
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

func (c AppsAchievementsQueueElastic) Group() TaskGroup {
	return TaskGroupElastic
}

func (c AppsAchievementsQueueElastic) Cron() TaskTime {
	return ""
}

func (c AppsAchievementsQueueElastic) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	var appCache = mongo.App{}

	for {

		// Sort by app id so we only need to cache one app at a time
		appAchievements, err := mongo.GetAppAchievements(offset, limit, nil, bson.D{{"app_id", 1}})
		if err != nil {
			return err
		}

		for _, appAchievement := range appAchievements {

			if appCache.ID != appAchievement.AppID {
				appCache, err = mongo.GetApp(appAchievement.AppID)
				if err != nil {
					return err
				}
			}

			err = queue.ProduceAchievementSearch(appAchievement, appCache.Name, appCache.Owners)
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
