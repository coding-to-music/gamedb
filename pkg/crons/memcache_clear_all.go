package crons

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

func (c MemcacheClearAll) Group() TaskGroup {
	return ""
}

func (c MemcacheClearAll) Cron() TaskTime {
	return ""
}

func (c MemcacheClearAll) work() (err error) {

	return memcache.Client().DeleteAll()
}
