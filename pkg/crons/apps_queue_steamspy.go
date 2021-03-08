package crons

import (
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type AppsQueueSteamSpy struct {
	BaseTask
}

func (c AppsQueueSteamSpy) ID() string {
	return "apps-queue-steamspy"
}

func (c AppsQueueSteamSpy) Name() string {
	return "Update all app SteamSpy data"
}

func (c AppsQueueSteamSpy) Group() TaskGroup {
	return TaskGroupApps
}

func (c AppsQueueSteamSpy) Cron() TaskTime {
	return CronTimeSteamSpy
}

func (c AppsQueueSteamSpy) work() (err error) {

	var filter = bson.D{
		{"owners", bson.M{"$gt": 0}}, // Just to keep requests to SteamSpy down a bit
	}
	var projection = bson.M{"_id": 1}

	var count int

	err = mongo.BatchApps(filter, projection, func(apps []mongo.App) {

		for _, app := range apps {

			err = consumers.ProduceAppSteamSpy(app.ID)
			if err != nil {
				log.ErrS(err)
				return
			}

			count++
		}
	})

	log.Info("Apps queued for SteamSpy", zap.Int("apps", count))

	return err
}
