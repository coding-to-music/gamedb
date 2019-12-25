package memcache

import (
	"strconv"

	"github.com/Jleagle/memcache-go/memcache"
	"github.com/Jleagle/steam-go/steam"
)

var (
	// Counts
	MemcacheAppsCount                 = memcache.Item{Key: "apps-count", Expiration: 86400}
	MemcacheAppsWithAchievementsCount = memcache.Item{Key: "apps-achievements-count", Expiration: 86400}
	MemcachePackagesCount             = memcache.Item{Key: "packages-count", Expiration: 86400}
	MemcacheBundlesCount              = memcache.Item{Key: "bundles-count", Expiration: 86400}
	MemcacheUpcomingAppsCount         = memcache.Item{Key: "upcoming-apps-count", Expiration: 86400}
	MemcacheNewReleaseAppsCount       = memcache.Item{Key: "newly-released-apps-count", Expiration: 86400}
	MemcacheUpcomingPackagesCount     = memcache.Item{Key: "upcoming-packages-count", Expiration: 86400}
	MemcachePlayersCount              = memcache.Item{Key: "players-count", Expiration: 86400 * 1}
	MemcacheSalesCount                = memcache.Item{Key: "sales-count", Expiration: 60 * 10}
	MemcachePricesCount               = memcache.Item{Key: "prices-count", Expiration: 86400 * 7}
	MemcacheMongoCount                = func(key string) memcache.Item { return memcache.Item{Key: "mongo-count-" + key, Expiration: 60 * 60} }
	MemcacheUserEventsCount           = func(userID int) memcache.Item { return memcache.Item{Key: "players-events-count-" + strconv.Itoa(userID), Expiration: 86400} }
	MemcachePatreonWebhooksCount      = func(userID int) memcache.Item { return memcache.Item{Key: "patreon-webhooks-count-" + strconv.Itoa(userID), Expiration: 86400} }

	// Apps Page Dropdowns
	MemcacheTagKeyNames       = memcache.Item{Key: "tag-key-names", Expiration: 86400 * 7}
	MemcacheCategoryKeyNames  = memcache.Item{Key: "category-key-names", Expiration: 86400 * 7}
	MemcacheGenreKeyNames     = memcache.Item{Key: "genre-key-names", Expiration: 86400 * 7}
	MemcachePublisherKeyNames = memcache.Item{Key: "publisher-key-names", Expiration: 86400 * 7}
	MemcacheDeveloperKeyNames = memcache.Item{Key: "developer-key-names", Expiration: 86400 * 7}

	// Rows
	MemcacheChange        = func(changeID int64) memcache.Item { return memcache.Item{Key: "change-" + strconv.FormatInt(changeID, 10), Expiration: 0} }
	MemcacheGroup         = func(id string) memcache.Item { return memcache.Item{Key: "group-" + id, Expiration: 60 * 30} } // 30 mins, cant be infinite as we need the 'updatedAt' field to be fairly upto date
	MemcachePackage       = func(id int) memcache.Item { return memcache.Item{Key: "package-" + strconv.Itoa(id), Expiration: 0} }
	MemcachePlayer        = func(id int64) memcache.Item { return memcache.Item{Key: "player-" + strconv.FormatInt(id, 10), Expiration: 0} }
	MemcacheConfigItem    = func(id string) memcache.Item { return memcache.Item{Key: "config-item-" + id, Expiration: 0} }
	MemcacheAppPlayersRow = func(appID int) memcache.Item { return memcache.Item{Key: "app-players-" + strconv.Itoa(appID), Expiration: 10 * 60} }
	MemcacheStatRowID     = func(c string, id int) memcache.Item { return memcache.Item{Key: c + "-stat-id-" + strconv.Itoa(id), Expiration: 60 * 60 * 24} }
	MemcacheStatRowName   = func(c string, name string) memcache.Item { return memcache.Item{Key: c + "-stat-name-" + name, Expiration: 60 * 60 * 24} }

	// Queue checks - 1 Hour timeout
	MemcacheAppInQueue     = func(appID int) memcache.Item { return memcache.Item{Key: "app-in-queue-" + strconv.Itoa(appID), Expiration: 60 * 60, Value: []byte("1")} }
	MemcacheBundleInQueue  = func(bundleID int) memcache.Item { return memcache.Item{Key: "bundle-in-queue-" + strconv.Itoa(bundleID), Expiration: 60 * 60, Value: []byte("1")} }
	MemcachePackageInQueue = func(packageID int) memcache.Item { return memcache.Item{Key: "package-in-queue-" + strconv.Itoa(packageID), Expiration: 60 * 60, Value: []byte("1")} }
	MemcachePlayerInQueue  = func(playerID int64) memcache.Item { return memcache.Item{Key: "profile-in-queue-" + strconv.FormatInt(playerID, 10), Expiration: 60 * 60, Value: []byte("1")} }
	MemcacheGroupInQueue   = func(groupID string) memcache.Item { return memcache.Item{Key: "group-in-queue-" + groupID, Expiration: 60 * 60, Value: []byte("1")} }

	// Home
	MemcacheHomePlayers = func(sort string) memcache.Item { return memcache.Item{Key: "home-players-" + sort, Expiration: 60 * 60 * 48} }

	// App Bits
	MemcacheAppTags       = func(appID int) memcache.Item { return memcache.Item{Key: "app-tags-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheAppCategories = func(appID int) memcache.Item { return memcache.Item{Key: "app-categories-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheAppGenres     = func(appID int) memcache.Item { return memcache.Item{Key: "app-genres-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheAppDemos      = func(appID int) memcache.Item { return memcache.Item{Key: "app-demos-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheAppRelated    = func(appID int) memcache.Item { return memcache.Item{Key: "app-related-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheAppDLC        = func(appID int) memcache.Item { return memcache.Item{Key: "app-dlcs-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheAppPublishers = func(appID int) memcache.Item { return memcache.Item{Key: "app-publishers-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheAppDevelopers = func(appID int) memcache.Item { return memcache.Item{Key: "app-developers-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheAppBundles    = func(appID int) memcache.Item { return memcache.Item{Key: "app-bundles-" + strconv.Itoa(appID), Expiration: 0} }

	// Package Bits
	MemcachePackageBundles = func(packageID int) memcache.Item { return memcache.Item{Key: "package-bundles-" + strconv.Itoa(packageID), Expiration: 0} }

	// Players
	MemcachePlayerLevels = memcache.Item{Key: "player-levels", Expiration: 60 * 60 * 24}

	// Other
	MemcacheQueues                   = memcache.Item{Key: "queues", Expiration: 10}
	MemcachePopularApps              = memcache.Item{Key: "popular-apps", Expiration: 60 * 3}
	MemcachePopularNewApps           = memcache.Item{Key: "popular-new-apps", Expiration: 60}
	MemcacheTrendingApps             = memcache.Item{Key: "trending-apps", Expiration: 60 * 10}
	MemcacheTotalCommits             = memcache.Item{Key: "total-commits", Expiration: 60 * 60}
	MemcacheStatsAppTypes            = func(code steam.ProductCC) memcache.Item { return memcache.Item{Key: "stats-app-types-" + string(code), Expiration: 60 * 60 * 25} }
	MemcacheUserByAPIKey             = func(key string) memcache.Item { return memcache.Item{Key: "user-level-by-key-" + key, Expiration: 10 * 60} }
	MemcacheUniquePlayerCountryCodes = memcache.Item{Key: "unique-player-country-codes", Expiration: 60 * 60 * 24 * 7}
	MemcacheUniquePlayerStateCodes   = func(c string) memcache.Item { return memcache.Item{Key: "unique-player-state-codes-" + c, Expiration: 60 * 60 * 24 * 7} }
	MemcacheUniqueSaleTypes          = memcache.Item{Key: "unique-sale-types", Expiration: 60 * 60 * 1}
)
