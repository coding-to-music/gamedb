package tasks

import (
	"github.com/gamedb/gamedb/pkg/helpers"
)

type ClearUpcomingCache struct {
	BaseTask
}

func (c ClearUpcomingCache) ID() string {
	return "clear-upcoming-apps-cache"
}

func (c ClearUpcomingCache) Name() string {
	return "Clear upcoming apps cache"
}

func (c ClearUpcomingCache) Cron() string {
	return CronTimeClearUpcomingCache
}

func (c ClearUpcomingCache) work() (err error) {

	err = helpers.RemoveKeyFromMemCacheViaPubSub(helpers.MemcacheUpcomingAppsCount.Key)
	if err != nil {
		return err
	}

	return helpers.RemoveKeyFromMemCacheViaPubSub(helpers.MemcacheUpcomingPackagesCount.Key)
}
