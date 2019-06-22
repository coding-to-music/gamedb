package helpers

import (
	"encoding/json"
	"strconv"

	"cloud.google.com/go/pubsub"
	"github.com/Jleagle/memcache-go/memcache"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
)

type MemcacheItem = memcache.Item

var (
	ErrCacheMiss   = memcache.ErrCacheMiss
	memcacheClient = memcache.New("game-db-", config.Config.MemcacheDSN.Get())
)

var (
	// Counts
	MemcacheAppsCount             = memcache.Item{Key: "apps-count", Expiration: 86400}
	MemcachePackagesCount         = memcache.Item{Key: "packages-count", Expiration: 86400}
	MemcacheBundlesCount          = memcache.Item{Key: "bundles-count", Expiration: 86400}
	MemcacheUpcomingAppsCount     = memcache.Item{Key: "upcoming-apps-count", Expiration: 86400}
	MemcacheNewReleaseAppsCount   = memcache.Item{Key: "newly-released-apps-count", Expiration: 86400}
	MemcacheUpcomingPackagesCount = memcache.Item{Key: "upcoming-packages-count", Expiration: 86400}
	MemcachePlayersCount          = memcache.Item{Key: "players-count", Expiration: 86400 * 1}
	MemcachePricesCount           = memcache.Item{Key: "prices-count", Expiration: 86400 * 7}
	MemcacheMongoCount            = func(key string) memcache.Item { return memcache.Item{Key: "mongo-count-" + key, Expiration: 60 * 60} }
	MemcacheUserEventsCount       = func(userID int) memcache.Item { return memcache.Item{Key: "players-events-count-" + strconv.Itoa(userID), Expiration: 86400} }
	MemcachePatreonWebhooksCount  = func(userID int) memcache.Item { return memcache.Item{Key: "patreon-webhooks-count-" + strconv.Itoa(userID), Expiration: 86400} }

	// Apps Page Dropdowns
	MemcacheTagKeyNames       = memcache.Item{Key: "tag-key-names", Expiration: 86400 * 7}
	MemcacheGenreKeyNames     = memcache.Item{Key: "genre-key-names", Expiration: 86400 * 7}
	MemcachePublisherKeyNames = memcache.Item{Key: "publisher-key-names", Expiration: 86400 * 7}
	MemcacheDeveloperKeyNames = memcache.Item{Key: "developer-key-names", Expiration: 86400 * 7}

	// Rows
	MemcacheChange        = func(changeID int64) memcache.Item { return memcache.Item{Key: "change-" + strconv.FormatInt(changeID, 10), Expiration: 0} }
	MemcacheGroup         = func(id string) memcache.Item { return memcache.Item{Key: "group-" + id, Expiration: 60 * 30} } // 30 mins, cant be infinite as we need the 'updatedAt' field to be fairly upto date
	MemcachePackage       = func(id int) memcache.Item { return memcache.Item{Key: "package-" + strconv.Itoa(id), Expiration: 0} }
	MemcacheConfigItem    = func(key string) memcache.Item { return memcache.Item{Key: "config-item-" + key, Expiration: 0} }
	MemcacheAppPlayersRow = func(appID int) memcache.Item { return memcache.Item{Key: "app-players-" + strconv.Itoa(appID), Expiration: 10 * 60} }
	MemcacheStatRowID     = func(c string, id int) memcache.Item { return memcache.Item{Key: c + "-stat-id-" + strconv.Itoa(id), Expiration: 60 * 60 * 24} }
	MemcacheStatRowName   = func(c string, name string) memcache.Item { return memcache.Item{Key: c + "-stat-name-" + name, Expiration: 60 * 60 * 24} }

	// Queue checks
	MemcacheGroupInQueue = func(groupID string) memcache.Item { return memcache.Item{Key: "group-in-queue-" + groupID, Expiration: 60 * 60 * 24, Value: []byte("1")} }

	// Other
	MemcacheQueues         = memcache.Item{Key: "queues", Expiration: 10}
	MemcachePopularApps    = memcache.Item{Key: "popular-apps", Expiration: 60 * 3}
	MemcachePopularNewApps = memcache.Item{Key: "popular-new-apps", Expiration: 60}
	MemcacheTrendingApps   = memcache.Item{Key: "trending-apps", Expiration: 60 * 10}
	MemcacheUserLevelByKey = func(key string) memcache.Item { return memcache.Item{Key: "user-level-by-key-" + key, Expiration: 10 * 60} }

	// Badges
	MemcacheBadgeMaxEvent     = func(id int) memcache.Item { return memcache.Item{Key: "badge-max-event-" + strconv.Itoa(id), Expiration: 60 * 60 * 25} }
	MemcacheBadgeMaxEventFoil = func(id int) memcache.Item { return memcache.Item{Key: "badge-max-event-foil-" + strconv.Itoa(id), Expiration: 60 * 60 * 25} }
	MemcacheBadgeMaxSpecial   = func(id int) memcache.Item { return memcache.Item{Key: "badge-max-special-" + strconv.Itoa(id), Expiration: 60 * 60 * 25} }
)

func GetMemcache() *memcache.Memcache {
	return memcacheClient
}

func ListenToPubSub() {

	mc := GetMemcache()

	err := PubSubSubscribe(PubSubMemcache, func(m *pubsub.Message) {

		var ids []string

		err := json.Unmarshal(m.Data, &ids)
		log.Err(err)

		for _, id := range ids {
			err = mc.Delete(id)
			err = IgnoreErrors(err, memcache.ErrCacheMiss)
			log.Err(err)
		}
	})
	log.Err(err)
}

//
func RemoveKeyFromMemCacheViaPubSub(keys ...string) (err error) {

	_, err = Publish(PubSubTopicMemcache, keys)
	return err
}
