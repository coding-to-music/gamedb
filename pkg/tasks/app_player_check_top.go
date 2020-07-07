package tasks

import (
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsPlayerCheckTop struct {
	BaseTask
}

func (c AppsPlayerCheckTop) ID() string {
	return "app-players-top"
}

func (c AppsPlayerCheckTop) Name() string {
	return "Check apps for players (Top)"
}

func (c AppsPlayerCheckTop) Cron() string {
	return CronTimeAppPlayersTop
}

const topAppPlayers = 10 // And up are top apps

func (c AppsPlayerCheckTop) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000
	var filter = bson.D{{"player_peak_week", bson.M{"$gte": topAppPlayers}}}

	for {

		apps, err := mongo.GetApps(offset, limit, bson.D{{"_id", 1}}, filter, bson.M{"_id": 1})
		if err != nil {
			return err
		}

		var ids []int
		for _, v := range apps {
			if v.ID > 0 { // This is just here to stop storing things on app 0, which we use to store steam stats on
				ids = append(ids, v.ID)
			}
		}

		var chunks = helpers.ChunkInts(ids, 20)

		for _, chunk := range chunks {
			err = queue.ProduceAppPlayersTop(queue.AppPlayerMessage{IDs: chunk})
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
