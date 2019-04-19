package main

import (
	"time"

	"github.com/gamedb/website/pkg/crons"
	"github.com/gamedb/website/pkg/helpers"
	"github.com/gamedb/website/pkg/log"
	"github.com/robfig/cron"
)

func main() {

	var err error

	c := cron.New()

	// Daily
	err = c.AddFunc("1 0 0 * * *", crons.ClearUpcomingCache)
	log.Critical(err)

	err = c.AddFunc("0 0 0 * * *", crons.CronRanks)
	log.Critical(err)

	err = c.AddFunc("0 0 1 * * *", crons.CronGenres)
	log.Critical(err)

	err = c.AddFunc("0 0 2 * * *", crons.CronTags)
	log.Critical(err)

	err = c.AddFunc("0 0 3 * * *", crons.CronPublishers)
	log.Critical(err)

	err = c.AddFunc("0 0 4 * * *", crons.CronDevelopers)
	log.Critical(err)

	err = c.AddFunc("0 0 5 * * *", crons.CronDonations)
	log.Critical(err)

	err = c.AddFunc("0 0 12 * * *", crons.Instagram)
	log.Critical(err)

	// Every 3 hours
	err = c.AddFunc("0 0 */3 * * *", crons.CronCheckForPlayers)
	log.Critical(err)

	// Every 6 hours
	err = c.AddFunc("0 0 */6 * * *", crons.AutoUpdateProfiles)
	log.Critical(err)

	c.Start()

	// Scan for app players after deploy
	go func() {
		time.Sleep(time.Minute)
		crons.CronCheckForPlayers()
	}()

	helpers.KeepAlive()
}
