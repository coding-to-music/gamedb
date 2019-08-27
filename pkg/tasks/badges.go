package tasks

import (
	"math/rand"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type SetBadgeCache struct {
}

func (c SetBadgeCache) ID() string {
	return "update-random-badge"
}

func (c SetBadgeCache) Name() string {
	return "Set a random badge cache"
}

func (c SetBadgeCache) Cron() string {
	return "0 */5 * * * *"
}

func (c SetBadgeCache) work() {

	var err error

	// Get a random badge
	badge := mongo.Badges[rand.Intn(len(mongo.Badges))]

	if badge.IsSpecial() {

		err = badge.SetSpecialMax()
		log.Err(err)

		time.Sleep(time.Second * 10)

		err = badge.SetSpecialPlayers()
		log.Err(err)

	} else {

		err = badge.SetEventMax()
		log.Err(err)

		time.Sleep(time.Second * 10)

		err = badge.SetEventMaxFoil()
		log.Err(err)

		time.Sleep(time.Second * 10)

		err = badge.SetEventPlayers()
		log.Err(err)
	}
}
