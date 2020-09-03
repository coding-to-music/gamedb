package tasks

import (
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
)

type StatsTask struct {
	BaseTask
}

func (c StatsTask) ID() string {
	return "update-stats"
}

func (c StatsTask) Name() string {
	return "Update stats"
}

func (c StatsTask) Group() TaskGroup {
	return ""
}

func (c StatsTask) Cron() TaskTime {
	return CronTimeStats
}

func (c StatsTask) work() (err error) {

	appsCount, err := mongo.CountDocuments(mongo.CollectionApps, nil, 0)
	if err != nil {
		return err
	}

	types := []mongo.StatsType{
		mongo.StatsTypeCategories,
		mongo.StatsTypeDevelopers,
		mongo.StatsTypeGenres,
		mongo.StatsTypePublishers,
		mongo.StatsTypeTags,
	}

	for _, t := range types {
		err := mongo.BatchStats(t, func(stats []mongo.Stat) {
			for _, stat := range stats {
				err = queue.ProduceStats(stat.Type, stat.ID, appsCount)
			}
		})
		if err != nil {
			return err
		}
	}

	return nil
}
