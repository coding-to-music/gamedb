package memcache

import (
	"strconv"

	"github.com/Jleagle/memcache-go"
	"github.com/Jleagle/steam-go/steamapi"
)

var (
	// SQL Counts
	MemcacheBundlesCount = memcache.Item{Key: "bundles-count", Expiration: 86400}
	MemcacheMongoCount   = func(key string) memcache.Item { return memcache.Item{Key: "mongo-count-" + key, Expiration: 60 * 60} }

	// Apps Page Dropdowns
	MemcacheTagKeyNames       = memcache.Item{Key: "tag-key-names", Expiration: 86400 * 7}
	MemcacheCategoryKeyNames  = memcache.Item{Key: "category-key-names", Expiration: 86400 * 7}
	MemcacheGenreKeyNames     = memcache.Item{Key: "genre-key-names", Expiration: 86400 * 7}
	MemcachePublisherKeyNames = memcache.Item{Key: "publisher-key-names", Expiration: 86400 * 7}
	MemcacheDeveloperKeyNames = memcache.Item{Key: "developer-key-names", Expiration: 86400 * 7}
	MemcacheAppTypeCounts     = memcache.Item{Key: "app-type-counts", Expiration: 86400 * 7}

	// Single Rows
	MemcacheApp        = func(id int) memcache.Item { return memcache.Item{Key: "app-" + strconv.Itoa(id), Expiration: 0} }
	MemcacheChange     = func(changeID int64) memcache.Item { return memcache.Item{Key: "change-" + strconv.FormatInt(changeID, 10), Expiration: 0} }
	MemcacheGroup      = func(id string) memcache.Item { return memcache.Item{Key: "group-" + id, Expiration: 0} }
	MemcachePackage    = func(id int) memcache.Item { return memcache.Item{Key: "package-" + strconv.Itoa(id), Expiration: 0} }
	MemcachePlayer     = func(id int64) memcache.Item { return memcache.Item{Key: "player-" + strconv.FormatInt(id, 10), Expiration: 0} }
	MemcacheConfigItem = func(id string) memcache.Item { return memcache.Item{Key: "config-item-" + id, Expiration: 0} }

	// Queue checks - 1 Hour timeout
	MemcacheAppInQueue     = func(appID int) memcache.Item { return memcache.Item{Key: "app-in-queue-" + strconv.Itoa(appID), Expiration: 60 * 60, Value: []byte("1")} }
	MemcacheBundleInQueue  = func(bundleID int) memcache.Item { return memcache.Item{Key: "bundle-in-queue-" + strconv.Itoa(bundleID), Expiration: 60 * 60, Value: []byte("1")} }
	MemcachePackageInQueue = func(packageID int) memcache.Item { return memcache.Item{Key: "package-in-queue-" + strconv.Itoa(packageID), Expiration: 60 * 60, Value: []byte("1")} }
	MemcachePlayerInQueue  = func(playerID int64) memcache.Item { return memcache.Item{Key: "profile-in-queue-" + strconv.FormatInt(playerID, 10), Expiration: 60 * 60, Value: []byte("1")} }
	MemcacheGroupInQueue   = func(groupID string) memcache.Item { return memcache.Item{Key: "group-in-queue-" + groupID, Expiration: 60 * 60, Value: []byte("1")} }

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
	MemcacheAppPackages   = func(appID int) memcache.Item { return memcache.Item{Key: "app-packages-" + strconv.Itoa(appID), Expiration: 0} }

	// Package Bits
	MemcachePackageBundles = func(packageID int) memcache.Item { return memcache.Item{Key: "package-bundles-" + strconv.Itoa(packageID), Expiration: 0} }

	// Home
	MemcacheHomePlayers = func(sort string) memcache.Item { return memcache.Item{Key: "home-players-" + sort, Expiration: 60 * 60 * 48} }

	// Queues page
	MemcacheQueues = memcache.Item{Key: "queues", Expiration: 10}

	// Chat page
	MemcacheChatBotGuilds = memcache.Item{Key: "chat-bot-guilds", Expiration: 60 * 60 * 24}

	// Sales page
	MemcacheUniqueSaleTypes = memcache.Item{Key: "unique-sale-types", Expiration: 60 * 60 * 1}

	// Players online
	MemcacheAppPlayersRow       = func(appID int) memcache.Item { return memcache.Item{Key: "app-players-" + strconv.Itoa(appID), Expiration: 10 * 60} }
	MemcacheAppPlayersInGameRow = memcache.Item{Key: "app-players-in-game-0", Expiration: 10 * 60}

	// Queries
	MemcachePopularApps    = memcache.Item{Key: "popular-apps", Expiration: 60 * 3}
	MemcachePopularNewApps = memcache.Item{Key: "popular-new-apps", Expiration: 60}
	MemcacheTrendingApps   = memcache.Item{Key: "trending-apps", Expiration: 60 * 10}

	// Other
	MemcacheTotalCommits             = memcache.Item{Key: "total-commits", Expiration: 60 * 60 * 24 * 7}
	MemcacheStatsAppTypes            = func(code steamapi.ProductCC) memcache.Item { return memcache.Item{Key: "stats-app-types-" + string(code), Expiration: 60 * 60 * 25} }
	MemcacheUserByAPIKey             = func(key string) memcache.Item { return memcache.Item{Key: "user-level-by-key-" + key, Expiration: 10 * 60} }
	MemcacheUniquePlayerCountryCodes = memcache.Item{Key: "unique-player-country-codes", Expiration: 60 * 60 * 24 * 7}
	MemcacheUniquePlayerStateCodes   = func(countryCode string) memcache.Item { return memcache.Item{Key: "unique-player-state-codes-" + countryCode, Expiration: 60 * 60 * 24 * 7} }
	MemcachePlayerLevels             = memcache.Item{Key: "player-levels", Expiration: 60 * 60 * 24}
)
