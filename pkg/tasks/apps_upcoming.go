package tasks

import (
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
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

	err = memcache.RemoveKeyFromMemCacheViaPubSub(memcache.MemcacheUpcomingAppsCount.Key)
	if err != nil {
		return err
	}

	return memcache.RemoveKeyFromMemCacheViaPubSub(memcache.MemcacheUpcomingPackagesCount.Key)
}
