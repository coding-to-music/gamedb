package memcache

import (
	"encoding/json"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/spf13/viper"
)

var client *memcache.Client
var ErrCacheMiss = memcache.ErrCacheMiss

var (
	// Counts
	AppsCount     = memcache.Item{Key: "apps-count", Expiration: 60 * 60 * 24}
	FreeAppsCount = memcache.Item{Key: "free-apps-count", Expiration: 60 * 60 * 24}
	PackagesCount = memcache.Item{Key: "packages-count", Expiration: 60 * 60 * 24}
	PlayersCount  = memcache.Item{Key: "players-count", Expiration: 60 * 60 * 24}

	// Full Tables
	Developers = memcache.Item{Key: "developers", Expiration: 60 * 60 * 24}
	Genres     = memcache.Item{Key: "genres", Expiration: 60 * 60 * 24}
	Publishers = memcache.Item{Key: "publishers", Expiration: 60 * 60 * 24}
	Tags       = memcache.Item{Key: "tags", Expiration: 60 * 60 * 24}
	FreeGames  = memcache.Item{Key: "free-games", Expiration: 0}
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

func GetSet(item memcache.Item, value interface{}, f func(j interface{}) (err error)) error {

	err := Get(item.Key, value)

	if err == ErrCacheMiss || (err != nil && err.Error() == "EOF") {

		err := f(value)
		if err != nil {
			return err
		}

		err = Set(item.Key, value, item.Expiration)
		if err != nil {
			return err
		}

		return nil
	}

	return err
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
