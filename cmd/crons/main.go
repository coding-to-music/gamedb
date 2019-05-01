package main

import (
	"time"

	"github.com/gamedb/gamedb/pkg/crons"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/robfig/cron"
)

func main() {

	log.Info("Starting crons")

	var err error

	c := cron.New()

	// Every 10 minutes
	err = c.AddFunc("0 */10 * * * *", crons.SteamPlayers)
	log.Critical(err)

	// Every 3 hours
	err = c.AddFunc("0 0 */3 * * *", crons.AppPlayers)
	log.Critical(err)

	// Every 6 hours
	err = c.AddFunc("0 0 */6 * * *", crons.AutoPlayerRefreshes)
	log.Critical(err)

	// Every 24 hours
	err = c.AddFunc("0 1 0 * * *", crons.ClearUpcomingCache)
	log.Critical(err)

	err = c.AddFunc("0 0 0 * * *", crons.PlayerRanks)
	log.Critical(err)

	err = c.AddFunc("0 0 1 * * *", crons.Genres)
	log.Critical(err)

	err = c.AddFunc("0 0 2 * * *", crons.Tags)
	log.Critical(err)

	err = c.AddFunc("0 0 3 * * *", crons.Publishers)
	log.Critical(err)

	err = c.AddFunc("0 0 4 * * *", crons.Developers)
	log.Critical(err)

	err = c.AddFunc("0 0 5 * * *", crons.Donations)
	log.Critical(err)

	err = c.AddFunc("0 0 12 * * *", crons.Instagram)
	log.Critical(err)

	c.Start()

	// Scan for app players after deploy
	go func() {
		time.Sleep(time.Minute)
		crons.AppPlayers()
	}()

	helpers.KeepAlive()
}
