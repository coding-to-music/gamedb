package main

import (
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
	err = c.AddFunc("0 */10 * * * *", crons.SteamClientPlayers{}.Work)
	log.Critical(err)

	// Every 5 hours
	err = c.AddFunc("0 0 */5 * * *", crons.AppPlayers{}.Work)
	log.Critical(err)

	err = c.AddFunc("0 0 */6 * * *", crons.AutoPlayerRefreshes{}.Work)
	log.Critical(err)

	// 12.00
	err = c.AddFunc("0 0 0 * * *", crons.Wishlists{}.Work)
	log.Critical(err)

	// 12.01
	err = c.AddFunc("0 1 0 * * *", crons.ClearUpcomingCache{}.Work)
	log.Critical(err)

	// 12.02
	err = c.AddFunc("0 2 0 * * *", crons.PlayerRanks{}.Work)
	log.Critical(err)

	// 01.00
	err = c.AddFunc("0 0 1 * * *", crons.Genres{}.Work)
	log.Critical(err)

	// 02.00
	err = c.AddFunc("0 0 2 * * *", crons.Tags{}.Work)
	log.Critical(err)

	// 03.00
	err = c.AddFunc("0 0 3 * * *", crons.Publishers{}.Work)
	log.Critical(err)

	// 04.00
	err = c.AddFunc("0 0 4 * * *", crons.Developers{}.Work)
	log.Critical(err)

	// 12.00
	err = c.AddFunc("0 0 12 * * *", crons.Instagram{}.Work)
	log.Critical(err)

	c.Start()

	helpers.KeepAlive()
}
