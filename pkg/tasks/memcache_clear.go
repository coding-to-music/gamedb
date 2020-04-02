package tasks

import (
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
)

type MemcacheClear struct {
	BaseTask
}

func (c MemcacheClear) ID() string {
	return "clear-memcache"
}

func (c MemcacheClear) Name() string {
	return "Clear Memcache"
}

func (c MemcacheClear) Cron() string {
	return ""
}

func (c MemcacheClear) work() (err error) {

	return memcache.DeleteAll()
}
