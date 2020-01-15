package tasks

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
)

type QueueAppGroups struct {
	BaseTask
}

func (c QueueAppGroups) ID() string {
	return "queue-app-groups"
}

func (c QueueAppGroups) Name() string {
	return "Queue app groups"
}

func (c QueueAppGroups) Cron() string {
	return CronQueueAppGroups
}

func (c QueueAppGroups) work() (err error) {

	db, err := sql.GetMySQLClient()
	if err != nil {
		return err
	}

	var apps []sql.App

	db = db.Select([]string{"group_id"})
	db = db.Where("group_id != ''")
	db = db.Find(&apps)

	if db.Error != nil {
		return db.Error
	}

	for _, app := range apps {

		err = queue.ProduceGroup(queue.GroupMessage{ID: app.GroupID})
		if err != nil {
			log.Err(err)
		}
	}

	//
	log.Info(strconv.Itoa(len(apps)) + " groups queued")

	return nil
}
