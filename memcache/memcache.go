package memcache

import (
	"encoding/json"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/spf13/viper"
	"github.com/steam-authority/steam-authority/logger"
)

var client *memcache.Client
var ErrCacheMiss = memcache.ErrCacheMiss

var (
	// Counts
	AppsCount     = memcache.Item{Key: "apps-count", Expiration: 60 * 60 * 24}
	FreeAppsCount = memcache.Item{Key: "free-apps-count", Expiration: 60 * 60 * 24}
	PackagesCount = memcache.Item{Key: "packages-count", Expiration: 60 * 60 * 24}
	RanksCount    = memcache.Item{Key: "ranks-count", Expiration: 60 * 60 * 24}
	PlayersCount  = memcache.Item{Key: "players-count", Expiration: 60 * 60 * 24}

	// Full Tables
	//Developers = memcache.Item{Key: "developers", Expiration: 60 * 60 * 24}
	//Genres     = memcache.Item{Key: "genres", Expiration: 60 * 60 * 24}
	//Publishers = memcache.Item{Key: "publishers", Expiration: 60 * 60 * 24}
	//Tags       = memcache.Item{Key: "tags", Expiration: 60 * 60 * 24}
	//FreeGames  = memcache.Item{Key: "free-games", Expiration: 0}
)

func getClient() *memcache.Client {

	if client == nil {
		client = memcache.New(viper.GetString("MEMCACHE_DSN"))
	}

	return client
}

func Get(key string, i interface{}) error {

	client := getClient()

	item, err := client.Get(key)
	if err != nil {
		return err
	}

	if len(item.Value) > 0 {
		err = json.Unmarshal(item.Value, i)
		if err != nil {
			return err
		}
	}

	return nil
}

func Set(key string, i interface{}, expiration int32) error {

	bytes, err := json.Marshal(i)
	if err != nil {
		return err
	}

	client := getClient()
	item := new(memcache.Item)
	item.Key = key
	item.Value = bytes
	item.Expiration = expiration

	return client.Set(item)
}

func GetSetInt(item memcache.Item, value *int, f func() (j int, err error)) (count int, err error) {

	err = Get(item.Key, value)

	if err != nil && (err == ErrCacheMiss || err.Error() == "EOF") {

		logger.Info("Loading " + item.Key + " from memcache.")

		count, err := f()
		if err != nil {
			return count, err
		}

		err = Set(item.Key, count, item.Expiration)
		if err != nil {
			return count, err
		}

		return count, nil
	}

	return *value, err
}

func Inc(key string) (err error) {

	client := getClient()
	_, err = client.Increment(key, 1)

	return err
}

func Dec(key string) (err error) {

	client := getClient()
	_, err = client.Decrement(key, 1)

	return err
}

// todo, add button to admin
func Wipe() (err error) {

	client := getClient()

	return client.DeleteAll()
}
