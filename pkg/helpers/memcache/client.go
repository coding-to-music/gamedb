package memcache

import (
	"github.com/Jleagle/memcache-go"
	"github.com/gamedb/gamedb/pkg/config"
)

var (
	ErrCacheMiss = memcache.ErrCacheMiss
	client       = memcache.New("game-db-", config.Config.MemcacheDSN.Get())
)

func GetClient() *memcache.Memcache {
	return client
}
