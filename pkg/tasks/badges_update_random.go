package tasks

import (
	"math/rand"
	"reflect"

	"github.com/gamedb/gamedb/pkg/mongo"
)

type SetBadgeCache struct {
	BaseTask
}

func (c SetBadgeCache) ID() string {
	return "update-random-badge"
}

func (c SetBadgeCache) Name() string {
	return "Set a random badge cache"
}

func (c SetBadgeCache) Cron() string {
	return CronTimeSetBadgeCache
}

func (c SetBadgeCache) work() (err error) {

	// Get random map key
	keys := reflect.ValueOf(mongo.GlobalBadges).MapKeys()
	randomID := keys[rand.Intn(len(keys))].Interface().(int)

	// Update random badge
	return mongo.UpdateBadgeSummary(randomID)
}
