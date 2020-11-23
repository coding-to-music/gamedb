package memcache

import (
	"strconv"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"go.mongodb.org/mongo-driver/bson"
)

type Item struct {
	Key        string // Key is the Item's key (250 bytes maximum).
	Value      string // Value is the Item's value.
	Flags      uint32 // Flags are server-opaque flags whose semantics are entirely up to the app.
	Expiration uint32 // Expiration is the cache expiration time, in seconds: either a relative time from now (up to 1 month), or an absolute Unix epoch time. Zero means the Item has no expiration time.
	// casid   uint64 // Compare and swap ID.
}

var (
	// App
	MemcacheAppReleaseDateCounts  = Item{Key: "app-release-date-counts", Expiration: 60 * 60 * 24}
	MemcacheAppReviewScoreCounts  = Item{Key: "app-review-score-counts", Expiration: 60 * 60 * 24 * 2}
	MemcacheApp                   = func(changeID int) Item { return Item{Key: "app-" + strconv.Itoa(changeID), Expiration: 0} }
	MemcacheAppTypeCounts         = func(cc steamapi.ProductCC) Item { return Item{Key: "app-type-counts-" + string(cc), Expiration: 86400 * 7} }
	MemcacheAppStats              = func(typex string, appID int) Item { return Item{Key: "app-stats-" + typex + "-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheAppDemos              = func(appID int) Item { return Item{Key: "app-demos-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheAppRelated            = func(appID int) Item { return Item{Key: "app-related-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheAppBundles            = func(appID int) Item { return Item{Key: "app-bundles-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheAppPackages           = func(appID int) Item { return Item{Key: "app-packages-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheAppNoAchievements     = func(appID int) Item { return Item{Key: "app-no-stats-" + strconv.Itoa(appID), Expiration: 60 * 60, Value: "1"} }
	MemcacheAppAchievementsCounts = func(appID int) Item { return Item{Key: "app-ach-counts-" + strconv.Itoa(appID), Expiration: 60 * 60 * 24} }

	MemcacheAppPlayersInGameRow    = Item{Key: "app-players-in-game-0", Expiration: 10 * 60}
	MemcacheAppPlayersRow          = func(appID int) Item { return Item{Key: "app-players-" + strconv.Itoa(appID), Expiration: 10 * 60} }
	MemcacheAppPlayersChart        = func(appID string, limited bool) Item { return Item{Key: "app-players-chart-" + appID + "-" + strconv.FormatBool(limited), Expiration: 10 * 60} }
	MemcacheAppPlayersHeatmapChart = func(appID string) Item { return Item{Key: "app-players-heatmap-chart-" + appID, Expiration: 10 * 60} }
	MemcacheAppTagsChart           = func(appID int) Item { return Item{Key: "app-tags-chart-" + strconv.Itoa(appID), Expiration: 10 * 60} }
	MemcacheAppWishlistChart       = func(appID string) Item { return Item{Key: "app-wishlist-chart-" + appID, Expiration: 10 * 60} }

	MemcachePopularApps    = Item{Key: "popular-apps", Expiration: 60 * 3}
	MemcachePopularNewApps = Item{Key: "popular-new-apps", Expiration: 60}
	MemcacheTrendingApps   = Item{Key: "trending-apps", Expiration: 60 * 10}

	// Bundle
	MemcacheBundlesCount = Item{Key: "bundles-count", Expiration: 86400}

	// Change
	MemcacheChange = func(changeID int64) Item { return Item{Key: "change-" + strconv.FormatInt(changeID, 10), Expiration: 0} }

	// Chat
	MemcacheChatBotGuildsCount = Item{Key: "chat-bot-guilds", Expiration: 60 * 60 * 24}
	MemcacheChatBotGuilds      = Item{Key: "chat-bot-grouped-guilds", Expiration: 60 * 10}
	MemcacheChatBotCommands    = Item{Key: "chat-bot-grouped-commands", Expiration: 60 * 10}
	MemcacheChatBotSettings    = func(discordID string) Item { return Item{Key: "chat-bot-settings-" + discordID, Expiration: 0} }
	MemcacheChatBotRequest     = func(request string, code steamapi.ProductCC) Item { return Item{Key: "chat-bot-request-" + string(code) + "-" + helpers.MD5([]byte(request)), Expiration: 60 * 10} }

	// GitHub
	MemcacheCommitsPage = func(page int) Item { return Item{Key: "commits-page-" + strconv.Itoa(page), Expiration: 60 * 60} }

	// Group
	MemcacheTrendingGroups      = Item{Key: "trending-groups", Expiration: 60 * 10}
	MemcacheGroup               = func(changeID string) Item { return Item{Key: "group-" + changeID, Expiration: 0} }
	MemcacheGroupFollowersChart = func(groupID string) Item { return Item{Key: "group-followers-chart-" + groupID, Expiration: 10 * 60} }

	// Package
	MemcachePackage        = func(changeID int) Item { return Item{Key: "package-" + strconv.Itoa(changeID), Expiration: 0} }
	MemcachePackageBundles = func(packageID int) Item { return Item{Key: "package-bundles-" + strconv.Itoa(packageID), Expiration: 0} }

	// Home
	HomeTweets          = Item{Key: "home-tweets", Expiration: 60 * 60 * 24 * 7}
	HomeTopSellers      = Item{Key: "home-top-sellers", Expiration: 60 * 60 * 6}
	MemcacheHomePlayers = func(sort string) Item { return Item{Key: "home-players-" + sort, Expiration: 60 * 60 * 48} }

	// Queue
	MemcacheQueues         = Item{Key: "queues", Expiration: 9} // Frontend refreshes every 10 seconds
	MemcacheAppInQueue     = func(appID int) Item { return Item{Key: "app-in-queue-" + strconv.Itoa(appID), Expiration: 60 * 60, Value: "1"} }
	MemcacheBundleInQueue  = func(bundleID int) Item { return Item{Key: "bundle-in-queue-" + strconv.Itoa(bundleID), Expiration: 60 * 60, Value: "1"} }
	MemcachePackageInQueue = func(packageID int) Item { return Item{Key: "package-in-queue-" + strconv.Itoa(packageID), Expiration: 60 * 60, Value: "1"} }
	MemcachePlayerInQueue  = func(playerID int64) Item { return Item{Key: "profile-in-queue-" + strconv.FormatInt(playerID, 10), Expiration: 60 * 60, Value: "1"} }
	MemcacheGroupInQueue   = func(groupID string) Item { return Item{Key: "group-in-queue-" + groupID, Expiration: 60 * 60, Value: "1"} }

	// Sales page
	MemcacheUniqueSaleTypes = Item{Key: "unique-sale-types", Expiration: 60 * 60 * 1}

	// Stat
	MemcacheStat           = func(t string, id int) Item { return Item{Key: "stat-" + t + "_" + strconv.Itoa(id), Expiration: 0} }
	MemcacheStatTime       = func(statKey string, cc steamapi.ProductCC) Item { return Item{Key: "stat-time-" + statKey + "-" + string(cc), Expiration: 60 * 60 * 6} }
	MemcacheStatsForSelect = func(t string) Item { return Item{Key: "stats-select-" + t, Expiration: 60 * 60 * 24} }

	// User
	MemcacheUserEvents    = func(userID int) Item { return Item{Key: "user-event-counts" + strconv.Itoa(userID), Expiration: 0} }
	MemcacheUserByAPIKey  = func(key string) Item { return Item{Key: "user-level-by-key-" + key, Expiration: 10 * 60} }
	MemcacheUserInDiscord = func(discordID string) Item { return Item{Key: "discord-id-" + discordID, Expiration: 60 * 60 * 24} }

	// Player
	MemcachePlayer                   = func(playerID int64) Item { return Item{Key: "player-" + strconv.FormatInt(playerID, 10), Expiration: 0} }
	MemcachePlayerAchievementsDays   = func(playerID int64) Item { return Item{Key: "player-ach-days-" + strconv.FormatInt(playerID, 10), Expiration: 0} }
	MemcachePlayerAchievementsInflux = func(playerID int64) Item { return Item{Key: "player-ach-influx-" + strconv.FormatInt(playerID, 10), Expiration: 0} }
	MemcachePlayerLevels             = Item{Key: "player-levels", Expiration: 60 * 60 * 24}
	MemcachePlayerUpdateDates        = Item{Key: "player-update-days", Expiration: 60 * 60 * 24}
	MemcachePlayerLevelsRounded      = Item{Key: "player-levels-rounded", Expiration: 60 * 60 * 24}
	MemcachePlayerLocationAggs       = Item{Key: "player-location-aggs", Expiration: 60 * 60 * 2}

	// Other
	MemcacheAPISteam      = Item{Key: "api-steam", Expiration: 60 * 60 * 24 * 7}
	MemcacheConfigItem    = func(configID string) Item { return Item{Key: "config-item-" + configID, Expiration: 0} }
	MemcacheFirstAppBadge = func(appID int) Item { return Item{Key: "first-app-badge-" + strconv.Itoa(appID), Expiration: 0} }
	MemcacheMongoCount    = func(collection string, filter bson.D) Item { return Item{Key: "mongo-count-" + collection + "-" + FilterToString(filter), Expiration: 60 * 60} }
)
