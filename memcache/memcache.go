package memcache

import (
	"encoding/json"
	"os"

	"github.com/bradfitz/gomemcache/memcache"
)

var client *memcache.Client
var ErrCacheMiss = memcache.ErrCacheMiss

var (
	// Counts
	AppsCount     = memcache.Item{Key: "apps-count", Expiration: 60 * 60 * 24}
	PackagesCount = memcache.Item{Key: "packages-count", Expiration: 60 * 60 * 24}
	PlayersCount  = memcache.Item{Key: "players-count", Expiration: 60 * 60 * 24}

	// Full Tables
	Developers = memcache.Item{Key: "developers", Expiration: 60 * 60 * 24}
	Genres     = memcache.Item{Key: "genres", Expiration: 60 * 60 * 24}
	Publishers = memcache.Item{Key: "publishers", Expiration: 60 * 60 * 24}
	Tags       = memcache.Item{Key: "tags", Expiration: 60 * 60 * 24}
)

func getClient() *memcache.Client {

	if client == nil {
		client = memcache.New(os.Getenv("STEAM_MEMCACHE"))
	}

	return client
}

func Get(key string, i interface{}) error {

	client := getClient()

	item, err := client.Get(key)
	if err != nil {
		return err
	}

	err = json.Unmarshal(item.Value, i)
	if err != nil {
		return err
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

func GetSet(key string, i interface{}, f func(j interface{}) (expiration int32, err error)) error {

	err := Get(key, i)

	if err == ErrCacheMiss || (err != nil && err.Error() == "EOF") {

		expiration, err := f(i)
		if err != nil {
			return err
		}

		err = Set(key, i, expiration)
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
