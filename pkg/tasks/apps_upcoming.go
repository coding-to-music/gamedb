package tasks

import (
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
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

func (c ClearUpcomingCache) work() {

	var err error

	err = helpers.RemoveKeyFromMemCacheViaPubSub(helpers.MemcacheUpcomingAppsCount.Key)
	log.Err(err)

	err = helpers.RemoveKeyFromMemCacheViaPubSub(helpers.MemcacheUpcomingPackagesCount.Key)
	log.Err(err)
}
