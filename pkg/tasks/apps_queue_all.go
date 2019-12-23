package tasks

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
)

type AppQueueAll struct {
	BaseTask
}

func (c AppQueueAll) ID() string {
	return "queue-all-apps"
}

func (c AppQueueAll) Name() string {
	return "Queue all apps"
}

func (c AppQueueAll) Cron() string {
	return ""
}

func (c AppQueueAll) work() (err error) {

	var last = 0
	var keepGoing = true
	var count int

	for keepGoing {

		apps, b, err := steam.GetSteam().GetAppList(1000, last, 0, "")
		err = steam.AllowSteamCodes(err, b, nil)
		if err != nil {
			return err
		}

		count = count + len(apps.Apps)

		for _, v := range apps.Apps {

			err = queue.ProduceApp(queue.AppMessage{ID: v.AppID})
			if err != nil {
				log.Err(err, strconv.Itoa(v.AppID))
				continue
			}
			last = v.AppID
		}

		keepGoing = apps.HaveMoreResults
	}

	log.Info("Found " + strconv.Itoa(count) + " apps")

	return nil
}
