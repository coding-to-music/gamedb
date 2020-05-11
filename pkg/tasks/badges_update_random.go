package tasks

import (
	"math/rand"
	"reflect"

	"github.com/gamedb/gamedb/pkg/mongo"
)

type BadgesUpdateRandom struct {
	BaseTask
}

func (c BadgesUpdateRandom) ID() string {
	return "badges-update-random"
}

func (c BadgesUpdateRandom) Name() string {
	return "Set a random badge cache"
}

func (c BadgesUpdateRandom) Cron() string {
	return CronTimeSetBadgeCache
}

func (c BadgesUpdateRandom) work() (err error) {

	// Get random map key
	keys := reflect.ValueOf(mongo.GlobalBadges).MapKeys()
	randomID := keys[rand.Intn(len(keys))].Interface().(int)

	// Update random badge
	return mongo.UpdateBadgeSummary(randomID)
}
