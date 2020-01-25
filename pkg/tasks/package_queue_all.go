package tasks

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type PackagesQueueAll struct {
	BaseTask
}

func (c PackagesQueueAll) ID() string {
	return "queue-all-packages"
}

func (c PackagesQueueAll) Name() string {
	return "Queue all packages"
}

func (c PackagesQueueAll) Cron() string {
	return ""
}

func (c PackagesQueueAll) work() (err error) {

	apps, err := mongo.GetNonEmptyArrays("packages", bson.M{"packages": 1})
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

	//
	log.Info(strconv.Itoa(len(packageMap)) + " packages added to rabbit")

	return nil
}
