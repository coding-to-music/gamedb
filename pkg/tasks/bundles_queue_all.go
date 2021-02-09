package tasks

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
)

type BundlesQueueAll struct {
	BaseTask
}

func (c BundlesQueueAll) ID() string {
	return "queue-all-bundles"
}

func (c BundlesQueueAll) Name() string {
	return "Queue all bundles"
}

func (c BundlesQueueAll) Group() TaskGroup {
	return TaskGroupBundles
}

func (c BundlesQueueAll) Cron() TaskTime {
	return ""
}

func (c BundlesQueueAll) work() (err error) {

	db, err := mysql.GetMySQLClient()
	if err != nil {
		log.ErrS(err)
		return
	}

	var bundles []mysql.Bundle

	db = db.Model(&mysql.Bundle{})
	db = db.Select([]string{"id"})
	db = db.Order("id asc")
	db = db.Find(&bundles)

	if db.Error != nil {
		return db.Error
	}

	for _, bundle := range bundles {

		err = queue.ProduceBundle(bundle.ID)
		if err != nil {
			log.ErrS(err)
			return
		}
	}

	return nil
}
