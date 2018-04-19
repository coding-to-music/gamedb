package memcache

import (
	"github.com/bradfitz/gomemcache/memcache"
)

var client *memcache.Client

func GetClient() *memcache.Client {

	if client == nil {
		client = memcache.New("localhost:11211") //todo make env var
	}

	return client
}
