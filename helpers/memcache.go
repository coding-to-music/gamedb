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
	MemcacheBundlesCount          = memcache.Item{Key: "bundles-count", Expiration: 86400}
	MemcacheUpcomingAppsCount     = memcache.Item{Key: "upcoming-apps-count", Expiration: 86400}
	MemcacheUpcomingPackagesCount = memcache.Item{Key: "upcoming-packages-count", Expiration: 86400}
	MemcacheRanksCount            = memcache.Item{Key: "ranks-count", Expiration: 86400}
	MemcachePlayersCount          = memcache.Item{Key: "players-count", Expiration: 86400 * 7}
	MemcachePlayerEventsCount     = func(playerID int64) memcache.Item {
		return memcache.Item{Key: "players-events-count-" + strconv.FormatInt(playerID, 10), Expiration: 86400}
	}

	// Dropdowns
	MemcacheTagKeyNames       = memcache.Item{Key: "tag-key-names", Expiration: 86400 * 7}
	MemcacheGenreKeyNames     = memcache.Item{Key: "genre-key-names", Expiration: 86400 * 7}
	MemcachePublisherKeyNames = memcache.Item{Key: "publisher-key-names", Expiration: 86400 * 7}
	MemcacheDeveloperKeyNames = memcache.Item{Key: "developer-key-names", Expiration: 86400 * 7}

	// Rows
	MemcacheChangeRow = func(changeID int64) memcache.Item {
		return memcache.Item{Key: "change-" + strconv.FormatInt(changeID, 10), Expiration: 86400 * 30}
	}
	MemcacheConfigRow = func(key string) memcache.Item {
		return memcache.Item{Key: "config-item-" + key, Expiration: 0}
	}
	MemcacheAppPlayersRow = func(appID int) memcache.Item {
		return memcache.Item{Key: "app-players-" + strconv.Itoa(appID), Expiration: 10 * 60}
	}

	// Other
	MemcacheMostExpensiveApp = func(code steam.CountryCode) memcache.Item {
		return memcache.Item{Key: "most-expensive-app-" + string(code), Expiration: 86400 * 7}
	}
	MemcacheQueues       = memcache.Item{Key: "queues", Expiration: 10}
	MemcachePopularApps  = memcache.Item{Key: "popular-apps", Expiration: 60 * 3}
	MemcacheTrendingApps = memcache.Item{Key: "trending-apps", Expiration: 60 * 3}
)
