package tasks

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/steam"
)

type AppsQueueAll struct {
	BaseTask
}

func (c AppsQueueAll) ID() string {
	return "queue-all-apps"
}

func (c AppsQueueAll) Name() string {
	return "Queue all apps"
}

func (c AppsQueueAll) Group() TaskGroup {
	return TaskGroupApps
}

func (c AppsQueueAll) Cron() TaskTime {
	return ""
}

func (c AppsQueueAll) work() (err error) {

	var last = 0
	var keepGoing = true
	var count int

	for keepGoing {

		apps, err := steam.GetSteam().GetAppList(1000, last, 0, "")
		err = steam.AllowSteamCodes(err)
		if err != nil {
			return err
		}

		count = count + len(apps.Apps)

		for _, v := range apps.Apps {

			err = queue.ProduceApp(queue.AppMessage{ID: v.AppID})
			if err != nil {
				return err
			}
			last = v.AppID
		}

		keepGoing = apps.HaveMoreResults
	}

	log.Info("Found " + strconv.Itoa(count) + " apps")

	return nil
}
