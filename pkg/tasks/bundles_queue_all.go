package tasks

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
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

	return mongo.BatchBundles(nil, bson.M{"_id": 1}, func(bundles []mongo.Bundle) {

		for _, bundle := range bundles {

			err = queue.ProduceBundle(bundle.ID)
			if err != nil {
				log.ErrS(err)
				return
			}
		}
	})
}
