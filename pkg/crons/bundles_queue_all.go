package crons

import (
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
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

			err = consumers.ProduceBundle(bundle.ID)
			if err != nil && err != consumers.ErrInQueue {
				log.ErrS(err)
				return
			}
		}
	})
}
