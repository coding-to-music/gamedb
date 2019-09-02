package tasks

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
)

type PackagesQueueAll struct {
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

func (c PackagesQueueAll) work() {

	apps, err := sql.GetAppsWithColumnDepth("packages", 2, []string{"packages"})
	if err != nil {
		log.Err(err)
		return
	}

	packageMap := map[int]bool{}
	for _, app := range apps {

		packagesIDs, err := app.GetPackageIDs()
		if err != nil {
			log.Err(app.ID, err)
			continue
		}

		for _, packageID := range packagesIDs {
			packageMap[packageID] = true
		}
	}

	// Make into slice again
	var packageSlice []int
	for k := range packageMap {
		packageSlice = append(packageSlice, k)
	}

	err = queue.ProduceToSteam(queue.SteamPayload{PackageIDs: packageSlice})
	log.Err(err)

	//
	log.Info(strconv.Itoa(len(packageMap)) + " packages added to rabbit")
}
