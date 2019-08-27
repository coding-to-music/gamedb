package tasks

import (
	"github.com/gamedb/gamedb/pkg/helpers"
)

type ClearUpcomingCache struct {
}

func (c ClearUpcomingCache) ID() string {
	return "clear-upcoming-apps-cache"
}

func (c ClearUpcomingCache) Name() string {
	return "Clear upcoming apps cache"
}

func (c ClearUpcomingCache) Cron() string {
	return "0 1 0 * * *"
}

func (c ClearUpcomingCache) work() {

	var err error

	err = helpers.RemoveKeyFromMemCacheViaPubSub(helpers.MemcacheUpcomingAppsCount.Key)
	cronLogErr(err)

	err = helpers.RemoveKeyFromMemCacheViaPubSub(helpers.MemcacheUpcomingPackagesCount.Key)
	cronLogErr(err)
}
