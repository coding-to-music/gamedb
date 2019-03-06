package cache

import (
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
)

func PopularApps() (apps []db.App, err error) {

	var item = helpers.MemcachePopularApps

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &apps, func() (interface{}, error) {

		gorm, err := db.GetMySQLClient()
		if err != nil {
			return apps, err
		}

		gorm = gorm.Select([]string{"id", "name", "icon", "player_peak_week"})
		gorm = gorm.Order("player_peak_week desc")
		gorm = gorm.Limit(15)
		gorm = gorm.Find(&apps)

		return apps, err
	})

	return apps, err
}
