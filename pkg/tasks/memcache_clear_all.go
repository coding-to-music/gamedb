package tasks

import (
	"github.com/gamedb/gamedb/pkg/memcache"
)

type MemcacheClearAll struct {
	BaseTask
}

func (c MemcacheClearAll) ID() string {
	return "clear-memcache"
}

func (c MemcacheClearAll) Name() string {
	return "Clear Memcache"
}

func (c MemcacheClearAll) Cron() string {
	return ""
}

func (c MemcacheClearAll) work() (err error) {

	return memcache.DeleteAll()
}
