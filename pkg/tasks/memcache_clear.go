package tasks

import (
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
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

func (c MemcacheClear) work() {

	err := helpers.GetMemcache().DeleteAll()
	log.Err(err)
}
