package helpers

import (
	"strconv"

	"github.com/Jleagle/memcache-go/memcache"
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/config"
)

var ErrCacheMiss = memcache.ErrCacheMiss

var memcacheClient = memcache.New("game-db-", config.Config.MemcacheDSN.Get())

func GetMemcache() *memcache.Memcache {
	return memcacheClient
}

var (
	// Counts
	MemcacheAppsCount             = memcache.Item{Key: "apps-count", Expiration: 86400}
	MemcachePackagesCount         = memcache.Item{Key: "packages-count", Expiration: 86400}
	MemcacheUpcomingAppsCount     = memcache.Item{Key: "upcoming-apps-count", Expiration: 86400}
	MemcacheUpcomingPackagesCount = memcache.Item{Key: "upcoming-packages-count", Expiration: 86400}
	MemcacheFreeAppsCount         = memcache.Item{Key: "free-apps-count", Expiration: 86400}
	MemcacheRanksCount            = memcache.Item{Key: "ranks-count", Expiration: 86400}
	MemcacheCountPlayers          = memcache.Item{Key: "players-count", Expiration: 86400 * 7}
	MemcachePlayerEventsCount     = func(playerID int64) memcache.Item {
		return memcache.Item{Key: "players-events-count-" + strconv.FormatInt(playerID, 10), Expiration: 86400}
	}

	// Dropdowns
	MemcacheTagKeyNames       = memcache.Item{Key: "tag-key-names", Expiration: 86400 * 7}
	MemcacheGenreKeyNames     = memcache.Item{Key: "genre-key-names", Expiration: 86400 * 7}
	MemcachePublisherKeyNames = memcache.Item{Key: "publisher-key-names", Expiration: 86400 * 7}
	MemcacheDeveloperKeyNames = memcache.Item{Key: "developer-key-names", Expiration: 86400 * 7}
	MemcacheAppTypes          = memcache.Item{Key: "app-types", Expiration: 86400 * 7}

	// Rows
	MemcacheChangeRow = func(changeID int64) memcache.Item {
		return memcache.Item{Key: "change-" + strconv.FormatInt(changeID, 10), Expiration: 86400 * 30}
	}
	MemcacheConfigRow = func(key string) memcache.Item {
		return memcache.Item{Key: "config-item-" + key, Expiration: 0}
	}

	// Queues
	//MemcachePlayerInQueue = func(playerID int64) memcache.Item {
	//	return memcache.Item{Key: "player-in-queue-" + strconv.FormatInt(playerID, 10), Expiration: 86400 * 7, Value: []byte("x")}
	//}
	//MemcacheAppInQueue = func(appID int) memcache.Item {
	//	return memcache.Item{Key: "app-in-queue-" + strconv.Itoa(appID), Expiration: 86400 * 7, Value: []byte("x")}
	//}
	//MemcachePackageInQueue = func(packageID int) memcache.Item {
	//	return memcache.Item{Key: "package-in-queue-" + strconv.Itoa(packageID), Expiration: 86400 * 7, Value: []byte("x")}
	//}

	// Other
	MemcacheMostExpensiveApp = func(code steam.CountryCode) memcache.Item {
		return memcache.Item{Key: "most-expensive-app-" + string(code), Expiration: 86400 * 7}
	}
	MemcacheQueues = memcache.Item{Key: "queues", Expiration: 10}
)
