package crons

import (
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type BundlesQueueElastic struct {
	BaseTask
}

func (c BundlesQueueElastic) ID() string {
	return "bundles-queue-elastic"
}

func (c BundlesQueueElastic) Name() string {
	return "Queue all bundles to Elastic"
}

func (c BundlesQueueElastic) Group() TaskGroup {
	return TaskGroupElastic
}

func (c BundlesQueueElastic) Cron() TaskTime {
	return ""
}

func (c BundlesQueueElastic) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		bundles, err := mongo.GetBundles(offset, limit, bson.D{{"_id", 1}}, nil, nil)
		if err != nil {
			return err
		}

		for _, bundle := range bundles {

			err = queue.ProduceBundleSearch(bundle)
			if err != nil {
				return err
			}
		}

		if int64(len(bundles)) != limit {
			break
		}

		offset += limit
	}

	return nil
}
