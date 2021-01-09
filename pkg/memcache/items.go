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
	ItemAppReleaseDateCounts  = Item{Key: "app-release-date-counts", Expiration: 60 * 60 * 24}
	ItemAppReviewScoreCounts  = Item{Key: "app-review-score-counts", Expiration: 60 * 60 * 24 * 2}
	ItemApp                   = func(changeID int) Item { return Item{Key: "app-" + strconv.Itoa(changeID), Expiration: 0} }
	ItemAppTypeCounts         = func(cc steamapi.ProductCC) Item { return Item{Key: "app-type-counts-" + string(cc), Expiration: 86400 * 7} }
	ItemAppStats              = func(typex string, appID int) Item { return Item{Key: "app-stats-" + typex + "-" + strconv.Itoa(appID), Expiration: 0} }
	ItemAppDemos              = func(appID int) Item { return Item{Key: "app-demos-" + strconv.Itoa(appID), Expiration: 0} }
	ItemAppRelated            = func(appID int) Item { return Item{Key: "app-related-" + strconv.Itoa(appID), Expiration: 0} }
	ItemAppBundles            = func(appID int) Item { return Item{Key: "app-bundles-" + strconv.Itoa(appID), Expiration: 0} }
	ItemAppPackages           = func(appID int) Item { return Item{Key: "app-packages-" + strconv.Itoa(appID), Expiration: 0} }
	ItemAppNoAchievements     = func(appID int) Item { return Item{Key: "app-no-stats-" + strconv.Itoa(appID), Expiration: 60 * 60, Value: "1"} }
	ItemAppAchievementsCounts = func(appID int) Item { return Item{Key: "app-ach-counts-" + strconv.Itoa(appID), Expiration: 60 * 60 * 24} }
	ItemAppTagsChart          = func(appID int) Item { return Item{Key: "app-tags-chart-" + strconv.Itoa(appID), Expiration: 10 * 60} }
	ItemAppWishlistChart      = func(appID string) Item { return Item{Key: "app-wishlist-chart-" + appID, Expiration: 10 * 60} }

	// Apps
	ItemAppsPopular    = Item{Key: "popular-apps", Expiration: 60 * 3}
	ItemNewPopularApps = Item{Key: "popular-new-apps", Expiration: 60}
	ItemAppsTrending   = Item{Key: "trending-apps", Expiration: 60 * 10}

	// App Player
	ItemAppPlayersInGameRow    = Item{Key: "app-players-in-game-0", Expiration: 10 * 60}
	ItemAppPlayersRow          = func(appID int) Item { return Item{Key: "app-players-" + strconv.Itoa(appID), Expiration: 10 * 60} }
	ItemAppPlayersChart        = func(appID string, limited bool) Item { return Item{Key: "app-players-chart-" + appID + "-" + strconv.FormatBool(limited), Expiration: 10 * 60} }
	ItemAppPlayersHeatmapChart = func(appID string) Item { return Item{Key: "app-players-heatmap-chart-" + appID, Expiration: 10 * 60} }

	// Chat
	ItemChatBotGuildsCount = Item{Key: "chat-bot-guilds", Expiration: 60 * 60 * 24}
	ItemChatBotSettings    = func(discordID string) Item { return Item{Key: "chat-bot-settings-" + discordID, Expiration: 0} }
	ItemChatBotRequest     = func(request string, code steamapi.ProductCC) Item { return Item{Key: "chat-bot-request-" + string(code) + "-" + helpers.MD5([]byte(request)), Expiration: 60 * 10} }
	ItemChatBotCommands    = Item{Key: "chat-bot-grouped-commands", Expiration: 60 * 10}
	ItemChatBotGuilds      = Item{Key: "chat-bot-grouped-guilds", Expiration: 60 * 10}

	// Group
	ItemGroup               = func(changeID string) Item { return Item{Key: "group-" + changeID, Expiration: 0} }
	ItemGroupsTrending      = Item{Key: "trending-groups", Expiration: 60 * 10}
	ItemGroupFollowersChart = func(groupID string) Item { return Item{Key: "group-followers-chart-" + groupID, Expiration: 10 * 60} }

	// Package
	ItemPackage        = func(changeID int) Item { return Item{Key: "package-" + strconv.Itoa(changeID), Expiration: 0} }
	ItemPackageBundles = func(packageID int) Item { return Item{Key: "package-bundles-" + strconv.Itoa(packageID), Expiration: 0} }

	// Home
	ItemHomeTweets     = Item{Key: "home-tweets", Expiration: 60 * 60 * 24 * 7}
	ItemHomeTopSellers = Item{Key: "home-top-sellers", Expiration: 60 * 60 * 6}
	ItemHomePlayers    = func(sort string) Item { return Item{Key: "home-players-" + sort, Expiration: 60 * 60 * 48} }

	// Queue
	ItemQueues         = Item{Key: "queues", Expiration: 9} // Frontend refreshes every 10 seconds
	ItemAppInQueue     = func(appID int) Item { return Item{Key: "app-in-queue-" + strconv.Itoa(appID), Expiration: 60 * 60, Value: "1"} }
	ItemBundleInQueue  = func(bundleID int) Item { return Item{Key: "bundle-in-queue-" + strconv.Itoa(bundleID), Expiration: 60 * 60, Value: "1"} }
	ItemPackageInQueue = func(packageID int) Item { return Item{Key: "package-in-queue-" + strconv.Itoa(packageID), Expiration: 60 * 60, Value: "1"} }
	ItemPlayerInQueue  = func(playerID int64) Item { return Item{Key: "profile-in-queue-" + strconv.FormatInt(playerID, 10), Expiration: 60 * 60, Value: "1"} }
	ItemGroupInQueue   = func(groupID string) Item { return Item{Key: "group-in-queue-" + groupID, Expiration: 60 * 60, Value: "1"} }

	// Stat
	ItemStat           = func(t string, id int) Item { return Item{Key: "stat-" + t + "_" + strconv.Itoa(id), Expiration: 0} }
	ItemStatTime       = func(statKey string, cc steamapi.ProductCC) Item { return Item{Key: "stat-time-" + statKey + "-" + string(cc), Expiration: 60 * 60 * 6} }
	ItemStatsForSelect = func(t string) Item { return Item{Key: "stats-select-" + t, Expiration: 60 * 60 * 24} }

	// User
	ItemUserEvents    = func(userID int) Item { return Item{Key: "user-event-counts" + strconv.Itoa(userID), Expiration: 0} }
	ItemUserByAPIKey  = func(key string) Item { return Item{Key: "user-level-by-key-" + key, Expiration: 10 * 60} }
	ItemUserInDiscord = func(discordID string) Item { return Item{Key: "discord-id-" + discordID, Expiration: 60 * 60 * 24} }

	// Player
	ItemPlayer                   = func(playerID int64) Item { return Item{Key: "player-" + strconv.FormatInt(playerID, 10), Expiration: 0} }
	ItemPlayerAchievementsDays   = func(playerID int64) Item { return Item{Key: "player-ach-days-" + strconv.FormatInt(playerID, 10), Expiration: 0} }
	ItemPlayerAchievementsInflux = func(playerID int64) Item { return Item{Key: "player-ach-influx-" + strconv.FormatInt(playerID, 10), Expiration: 0} }
	ItemPlayerFriends            = func(playerID int64, appID int) Item { return Item{Key: "player-friends-" + strconv.FormatInt(playerID, 10) + "-" + strconv.Itoa(appID), Expiration: 60 * 60 * 24} }
	ItemPlayerLevels             = Item{Key: "player-levels", Expiration: 60 * 60 * 24}
	ItemPlayerLevelsRounded      = Item{Key: "player-levels-rounded", Expiration: 60 * 60 * 24}
	ItemPlayerLocationAggs       = Item{Key: "player-location-aggs", Expiration: 60 * 60 * 2}
	ItemPlayerUpdateDates        = Item{Key: "player-update-days", Expiration: 60 * 60 * 24}

	// Other
	ItemAPISteam             = Item{Key: "api-steam", Expiration: 60 * 60 * 24 * 7}
	ItemArticleFeedAggs      = Item{Key: "app-article-feeds", Expiration: 60 * 60 * 24 * 7}
	ItemArticleFeedAggsMongo = Item{Key: "app-article-feeds-mongo", Expiration: 60 * 60 * 24 * 7}
	ItemBundlesCount         = Item{Key: "bundles-count", Expiration: 86400}
	ItemChange               = func(changeID int64) Item { return Item{Key: "change-" + strconv.FormatInt(changeID, 10), Expiration: 0} }
	ItemCommitsPage          = func(page int) Item { return Item{Key: "commits-page-" + strconv.Itoa(page), Expiration: 60 * 60} }
	ItemConfigItem           = func(configID string) Item { return Item{Key: "config-item-" + configID, Expiration: 0} }
	ItemFirstAppBadge        = func(appID int) Item { return Item{Key: "first-app-badge-" + strconv.Itoa(appID), Expiration: 0} }
	ItemMongoCount           = func(collection string, filter bson.D) Item { return Item{Key: "mongo-count-" + collection + "-" + FilterToString(filter), Expiration: 60 * 60} }
	ItemUniqueSaleTypes      = Item{Key: "unique-sale-types", Expiration: 60 * 60 * 1}
)
