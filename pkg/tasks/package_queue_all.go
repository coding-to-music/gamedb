package tasks

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
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

	apps, err := sql.GetAppsWithColumnDepth("packages", 2, []string{"packages"})
	if err != nil {
		return err
	}

	packageMap := map[int]bool{}
	for _, app := range apps {
		for _, packageID := range app.GetPackageIDs() {
			packageMap[packageID] = true
		}
	}

	// Make into slice again
	var packageSlice []int
	for k := range packageMap {
		packageSlice = append(packageSlice, k)
	}

	err = queue.ProduceToSteam(queue.SteamPayload{PackageIDs: packageSlice, Force: true})
	if err != nil {
		return err
	}

	//
	log.Info(strconv.Itoa(len(packageMap)) + " packages added to rabbit")

	return nil
}
