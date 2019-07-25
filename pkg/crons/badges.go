package crons

import (
	"math/rand"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
)

type SetBadgeCache struct {
}

func (c SetBadgeCache) ID() CronEnum {
	return CronBadgesCache
}

func (c SetBadgeCache) Name() string {
	return "Set a random badge cache"
}

func (c SetBadgeCache) Config() sql.ConfigType {
	return sql.ConfBadgeCache
}

func (c SetBadgeCache) Work() {

	started(c)

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

	finished(c)
}
