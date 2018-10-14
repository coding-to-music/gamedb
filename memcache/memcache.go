package memcache

import (
	"encoding/json"
	"strconv"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/spf13/viper"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logger"
)

const namespace = "game-db-"

var client *memcache.Client
var ErrCacheMiss = memcache.ErrCacheMiss

var (
	// Counts
	AppsCount         = memcache.Item{Key: "apps-count", Expiration: 60 * 60 * 24}
	FreeAppsCount     = memcache.Item{Key: "free-apps-count", Expiration: 60 * 60 * 24}
	PackagesCount     = memcache.Item{Key: "packages-count", Expiration: 60 * 60 * 24}
	RanksCount        = memcache.Item{Key: "ranks-count", Expiration: 60 * 60 * 24}
	CountPlayers      = memcache.Item{Key: "players-count", Expiration: 60 * 60 * 24 * 7}
	PlayerEventsCount = func(playerID int64) memcache.Item {
		return memcache.Item{Key: "players-events-count-" + strconv.FormatInt(playerID, 10), Expiration: 60 * 60 * 24}
	}

	// Dropdowns
	TagKeyNames       = memcache.Item{Key: "tag-key-names", Expiration: 60 * 60 * 24 * 1}
	GenreKeyNames     = memcache.Item{Key: "genre-key-names", Expiration: 60 * 60 * 24 * 1}
	PublisherKeyNames = memcache.Item{Key: "publisher-key-names", Expiration: 60 * 60 * 24 * 1}
	DeveloperKeyNames = memcache.Item{Key: "developer-key-names", Expiration: 60 * 60 * 24 * 1}
	AppTypes          = memcache.Item{Key: "app-types", Expiration: 60 * 60 * 24 * 1}
)

func getClient() *memcache.Client {

	if client == nil {
		client = memcache.New(viper.GetString("MEMCACHE_DSN"))
	}

	return client
}

func Get(key string, i interface{}) error {

	client := getClient()

	item, err := client.Get(namespace + key)
	if err != nil {
		return err
	}

	return helpers.Unmarshal(item.Value, i)
}

func Set(key string, i interface{}, expiration int32) error {

	bytes, err := json.Marshal(i)
	if err != nil {
		return err
	}

	client := getClient()
	item := new(memcache.Item)
	item.Key = namespace + key
	item.Value = bytes
	item.Expiration = expiration

	return client.Set(item)
}

func GetSetInt(item memcache.Item, f func() (j int, err error)) (count int, err error) {

	err = Get(item.Key, &count)

	if err != nil && (err == ErrCacheMiss || err.Error() == "EOF") {

		logger.Info("Loading " + item.Key + " from memcache.")

		count, err := f()
		if err != nil {
			return count, err
		}

		err = Set(item.Key, count, item.Expiration)
		return count, err
	}

	return count, err
}

func GetSetString(item memcache.Item, f func() (j string, err error)) (s string, err error) {

	err = Get(item.Key, &s)

	if err != nil && (err == ErrCacheMiss || err.Error() == "EOF") {

		logger.Info("Loading " + item.Key + " from memcache.")

		s, err := f()
		if err != nil {
			return s, err
		}

		err = Set(item.Key, s, item.Expiration)
		return s, err
	}

	return s, err
}

func Inc(key string) (err error) {

	client := getClient()
	_, err = client.Increment(namespace+key, 1)

	return err
}

func Dec(key string) (err error) {

	client := getClient()
	_, err = client.Decrement(namespace+key, 1)

	return err
}

// todo, add button to admin
func Wipe() (err error) {

	client := getClient()

	return client.DeleteAll()
}
