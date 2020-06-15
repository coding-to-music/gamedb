package memcache

import (
	"strconv"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
)

type Item struct {
	Key        string // Key is the Item's key (250 bytes maximum).
	Value      string // Value is the Item's value.
	Flags      uint32 // Flags are server-opaque flags whose semantics are entirely up to the app.
	Expiration uint32 // Expiration is the cache expiration time, in seconds: either a relative time from now (up to 1 month), or an absolute Unix epoch time. Zero means the Item has no expiration time.
	// casid   uint64 // Compare and swap ID.
}

var (
	// Counts
	MemcacheBundlesCount = Item{Key: "bundles-count", Expiration: 86400}
	MemcacheMongoCount   = func(key string) Item { return Item{Key: "mongo-count-" + key, Expiration: 60 * 60} }

	// Apps Page Dropdowns
	MemcacheTagKeyNames          = Item{Key: "tag-key-names", Expiration: 86400 * 7}
	MemcacheCategoryKeyNames     = Item{Key: "category-key-names", Expiration: 86400 * 7}
	MemcacheGenreKeyNames        = Item{Key: "genre-key-names", Expiration: 86400 * 7}
	MemcachePublisherKeyNames    = Item{Key: "publisher-key-names", Expiration: 86400 * 7}
	MemcacheDeveloperKeyNames    = Item{Key: "developer-key-names", Expiration: 86400 * 7}
	MemcacheAppTypeCounts        = func(cc steamapi.ProductCC) Item { return Item{Key: "app-type-counts-" + string(cc), Expiration: 86400 * 7} }
	MemcacheAppReleaseDateCounts = Item{Key: "app-release-date-counts", Expiration: 60 * 60 * 24}
	MemcacheAppReviewScoreCounts = Item{Key: "app-review-score-counts", Expiration: 60 * 60 * 24 * 2}

	// Single Rows
	MemcacheApp        = func(id int) Item { return Item{Key: "app-" + strconv.Itoa(id), Expiration: 0} }
	MemcacheChange     = func(changeID int64) Item { return Item{Key: "change-" + strconv.FormatInt(changeID, 10), Expiration: 0} }
	MemcacheGroup      = func(id string) Item { return Item{Key: "group-" + id, Expiration: 0} }
	MemcachePackage    = func(id int) Item { return Item{Key: "package-" + strconv.Itoa(id), Expiration: 0} }
	MemcachePlayer     = func(id int64) Item { return Item{Key: "player-" + strconv.FormatInt(id, 10), Expiration: 0} }
	MemcacheConfigItem = func(id string) Item { return Item{Key: "config-item-" + id, Expiration: 0} }

	// Queue checks - 1 Hour timeout
	MemcacheAppInQueue     = func(appID int) Item { return Item{Key: "app-in-queue-" + strconv.Itoa(appID), Expiration: 60 * 60, Value: "1"} }
	MemcacheBundleInQueue  = func(bundleID int) Item { return Item{Key: "bundle-in-queue-" + strconv.Itoa(bundleID), Expiration: 60 * 60, Value: "1"} }
	MemcachePackageInQueue = func(packageID int) Item { return Item{Key: "package-in-queue-" + strconv.Itoa(packageID), Expiration: 60 * 60, Value: "1"} }
	MemcachePlayerInQueue  = func(playerID int64) Item { return Item{Key: "profile-in-queue-" + strconv.FormatInt(playerID, 10), Expiration: 60 * 60, Value: "1"} }
	MemcacheGroupInQueue   = func(groupID string) Item { return Item{Key: "group-in-queue-" + groupID, Expiration: 60 * 60, Value: "1"} }

	// App Bits
	MemcacheAppTags       = func(appID int) Item { return Item{Key: "app-tags-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheAppCategories = func(appID int) Item { return Item{Key: "app-categories-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheAppGenres     = func(appID int) Item { return Item{Key: "app-genres-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheAppDemos      = func(appID int) Item { return Item{Key: "app-demos-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheAppRelated    = func(appID int) Item { return Item{Key: "app-related-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheAppPublishers = func(appID int) Item { return Item{Key: "app-publishers-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheAppDevelopers = func(appID int) Item { return Item{Key: "app-developers-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheAppBundles    = func(appID int) Item { return Item{Key: "app-bundles-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheAppPackages   = func(appID int) Item { return Item{Key: "app-packages-" + strconv.Itoa(appID), Expiration: 0} }

	// Groups
	MemcacheTrendingGroups = Item{Key: "trending-apps", Expiration: 60 * 10}

	// Chat Bot
	MemcacheChatBotGuildsCount = Item{Key: "chat-bot-guilds", Expiration: 60 * 60 * 24}
	MemcacheChatBotGuilds      = Item{Key: "chat-bot-grouped-guilds", Expiration: 60 * 10}
	MemcacheChatBotCommands    = Item{Key: "chat-bot-grouped-commands", Expiration: 60 * 10}

	// Package Bits
	MemcachePackageBundles = func(packageID int) Item { return Item{Key: "package-bundles-" + strconv.Itoa(packageID), Expiration: 0} }

	// Home
	MemcacheHomePlayers = func(sort string) Item { return Item{Key: "home-players-" + sort, Expiration: 60 * 60 * 48} }

	// Queues page
	MemcacheQueues = Item{Key: "queues", Expiration: 10}

	// Sales page
	MemcacheUniqueSaleTypes = Item{Key: "unique-sale-types", Expiration: 60 * 60 * 1}

	// Commits page
	MemcacheCommitsTotal = Item{Key: "commits-total", Expiration: 60 * 60}
	MemcacheCommitsPage  = func(page int) Item { return Item{Key: "commits-page-" + strconv.Itoa(page), Expiration: 60 * 60} }

	// Players online
	MemcacheAppPlayersRow       = func(appID int) Item { return Item{Key: "app-players-" + strconv.Itoa(appID), Expiration: 10 * 60} }
	MemcacheAppPlayersInGameRow = Item{Key: "app-players-in-game-0", Expiration: 10 * 60}

	// Queries
	MemcachePopularApps    = Item{Key: "popular-apps", Expiration: 60 * 3}
	MemcachePopularNewApps = Item{Key: "popular-new-apps", Expiration: 60}
	MemcacheTrendingApps   = Item{Key: "trending-apps", Expiration: 60 * 10}

	// Other
	MemcacheUserByAPIKey             = func(key string) Item { return Item{Key: "user-level-by-key-" + key, Expiration: 10 * 60} }
	MemcacheUniquePlayerCountryCodes = Item{Key: "unique-player-country-codes", Expiration: 60 * 60 * 24 * 7}
	MemcacheUniquePlayerStateCodes   = func(countryCode string) Item { return Item{Key: "unique-player-state-codes-" + countryCode, Expiration: 60 * 60 * 24 * 7} }
	MemcachePlayerLevels             = Item{Key: "player-levels", Expiration: 60 * 60 * 24}
	MemcachePlayerLevelsRounded      = Item{Key: "player-levels-rounded", Expiration: 60 * 60 * 24}
	MemcacheFirstAppBadge            = func(appID int) Item { return Item{Key: "first-app-badge-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheChatBotRequest           = func(request string) Item { return Item{Key: "chat-bot-request-" + helpers.MD5([]byte(request)), Expiration: 60 * 10} }
)
