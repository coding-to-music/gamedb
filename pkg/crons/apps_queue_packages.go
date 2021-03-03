package crons

import (
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsQueuePackages struct {
	BaseTask
}

func (c AppsQueuePackages) ID() string {
	return "queue-all-packages"
}

func (c AppsQueuePackages) Name() string {
	return "Queue all packages"
}

func (c AppsQueuePackages) Group() TaskGroup {
	return TaskGroupPackages
}

func (c AppsQueuePackages) Cron() TaskTime {
	return ""
}

func (c AppsQueuePackages) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		sort := bson.D{{"_id", 1}}
		projection := bson.M{"packages": 1}
		filter := bson.D{
			{"packages", bson.M{"$exists": true}},
			{"packages", bson.M{"$ne": bson.A{}}},
		}

		apps, err := mongo.GetApps(offset, limit, sort, filter, projection)
		if err != nil {
			return err
		}

		packageMap := map[int]bool{}
		for _, app := range apps {
			for _, packageID := range app.Packages {
				packageMap[packageID] = true
			}
		}

		// Make into slice again
		var packageSlice []int
		for k := range packageMap {
			packageSlice = append(packageSlice, k)
		}

		err = queue.ProduceSteam(queue.SteamMessage{PackageIDs: packageSlice})
		if err != nil {
			return err
		}

		if int64(len(apps)) != limit {
			break
		}

		offset += limit
	}

	return nil
}
