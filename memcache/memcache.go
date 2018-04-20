package memcache

import (
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

func GetClient() *memcache.Client {

	if client == nil {
		client = memcache.New(os.Getenv("STEAM_MEMCACHE"))
	}

	return client
}

func Set(key string, fn func() []byte) {

}
