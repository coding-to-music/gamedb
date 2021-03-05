package memcache

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Jleagle/memcache-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
)

type Item struct {
	Key        string // Key is the Item's key (250 bytes maximum).
	Value      string // Value is the Item's value.
	Expiration uint32 // Expiration is the cache expiration time, in seconds: either a relative time from now (up to 1 month), or an absolute Unix epoch time. Zero means the Item has no expiration time.
}

var (
	// App
	ItemAppReleaseDateCounts  = Item{Key: "app-release-date-counts", Expiration: 60 * 60 * 24}
	ItemAppReviewScoreCounts  = Item{Key: "app-review-score-counts", Expiration: 60 * 60 * 24 * 2}
	ItemApp                   = func(appID int) Item { return Item{Key: "app-" + strconv.Itoa(appID), Expiration: 0} }
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
	ItemAppTopPrice           = func(cc steamapi.ProductCC) Item { return Item{Key: "app-top-price-" + string(cc), Expiration: 86400} }

	// Apps
	ItemAppsPopular    = Item{Key: "popular-apps", Expiration: 60 * 3}
	ItemNewPopularApps = Item{Key: "popular-new-apps", Expiration: 60}
	ItemAppsTrending   = Item{Key: "trending-apps", Expiration: 60 * 10}

	// App Player
	ItemAppPlayersInGameRow    = Item{Key: "app-players-in-game-0", Expiration: 10 * 60}
	ItemAppPlayersRow          = func(appID int) Item { return Item{Key: "app-players-" + strconv.Itoa(appID), Expiration: 10 * 60} }
	ItemAppPlayersChart        = func(appID string, limited bool) Item { return Item{Key: "app-players-chart-" + appID + "-" + strconv.FormatBool(limited), Expiration: 10 * 60} }
	ItemAppPlayersHeatmapChart = func(appID string) Item { return Item{Key: "app-players-heatmap-chart-" + appID, Expiration: 10 * 60} }

	// Bundles
	ItemBundle = func(bundleID int) Item { return Item{Key: "bundle-" + strconv.Itoa(bundleID), Expiration: 0} }

	// Chat
	ItemChatBotSettings     = func(discordID string) Item { return Item{Key: "chat-bot-settings-" + discordID, Expiration: 0} }
	ItemChatBotRequest      = func(request string, code steamapi.ProductCC) Item { return Item{Key: "interaction-" + string(code) + "-" + helpers.MD5([]byte(request)), Expiration: 60 * 10} }
	ItemChatBotRequestSlash = func(commandID string, inputs map[string]string, code steamapi.ProductCC) Item { return Item{Key: "interaction-slash-" + commandID + "-" + string(code) + "-" + helpers.MD5Interface(inputs), Expiration: 60 * 10} }

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
	ItemHomeUpcoming   = Item{Key: "home-upcoming", Expiration: 60 * 60}
	ItemHomeNews       = Item{Key: "home-news", Expiration: 60 * 60}

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
	ItemArticleFeedAggsMongo = func(appID int) Item { return Item{Key: "app-article-feeds-mongo-" + strconv.Itoa(appID), Expiration: 60 * 60 * 24 * 7} }
	ItemChange               = func(changeID int64) Item { return Item{Key: "change-" + strconv.FormatInt(changeID, 10), Expiration: 0} }
	ItemCommitsPage          = func(page int) Item { return Item{Key: "commits-page-" + strconv.Itoa(page), Expiration: 60 * 60} }
	ItemConfigItem           = func(configID string) Item { return Item{Key: "config-item-" + configID, Expiration: 0} }
	ItemFirstAppBadge        = func(appID int) Item { return Item{Key: "first-app-badge-" + strconv.Itoa(appID), Expiration: 0} }
	ItemMongoCount           = func(collection string, filter bson.D) Item { return Item{Key: "mongo-count-" + collection + "-" + FilterToString(filter), Expiration: 60 * 60} }
	ItemUniqueSaleTypes      = Item{Key: "unique-sale-types", Expiration: 60 * 60 * 1}
	ItemChatbotCalls         = Item{Key: "chatbot-calls", Expiration: 60 * 10}
)

var lock sync.Mutex
var client *memcache.Client

func Client() *memcache.Client {

	lock.Lock()
	defer lock.Unlock()

	if client == nil {

		if config.C.MemcacheDSN == "" {
			log.Err("Missing environment variables")
		}

		options := []memcache.Option{
			memcache.WithAuth(config.C.MemcacheUsername, config.C.MemcachePassword),
			memcache.WithNamespace("gs_"),
		}

		if config.IsLocal() {
			options = append(options, memcache.WithTypeChecks(true))
		}

		client = memcache.NewClient(config.C.MemcacheDSN, options...)
	}

	return client
}

func FilterToString(d bson.D) string {

	if d == nil || len(d) == 0 {
		return "[]"
	}

	b, err := json.Marshal(d)
	if err != nil {
		log.ErrS(err)
		return "[]"
	}

	return helpers.MD5(b)
}

func ProjectionToString(m bson.M) string {

	if len(m) == 0 {
		return "*"
	}

	var cols []string
	for k := range m {
		cols = append(cols, k)
	}

	sort.Slice(cols, func(i, j int) bool {
		return cols[i] < cols[j]
	})

	return strings.Join(cols, "-")
}
