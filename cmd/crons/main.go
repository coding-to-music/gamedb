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
	err = c.AddFunc("0 */10 * * * *", crons.CronRegister[crons.CronSteamClientPlayers].Work)
	log.Critical(err)

	// Every 6 hours
	err = c.AddFunc("0 0 */6 * * *", crons.CronRegister[crons.CronAppPlayers].Work)
	log.Critical(err)

	err = c.AddFunc("0 0 */6 * * *", crons.CronRegister[crons.CronAutoPlayerRefreshes].Work)
	log.Critical(err)

	// 12.00
	err = c.AddFunc("0 0 0 * * *", crons.CronRegister[crons.CronWishlist].Work)
	log.Critical(err)

	// 12.01
	err = c.AddFunc("0 1 0 * * *", crons.CronRegister[crons.CronClearUpcomingCache].Work)
	log.Critical(err)

	// 12.02
	err = c.AddFunc("0 2 0 * * *", crons.CronRegister[crons.CronPlayerRanks].Work)
	log.Critical(err)

	// 01.00
	err = c.AddFunc("0 0 1 * * *", crons.CronRegister[crons.CronGenres].Work)
	log.Critical(err)

	// 02.00
	err = c.AddFunc("0 0 2 * * *", crons.CronRegister[crons.CronTags].Work)
	log.Critical(err)

	// 03.00
	err = c.AddFunc("0 0 3 * * *", crons.CronRegister[crons.CronPublishers].Work)
	log.Critical(err)

	// 04.00
	err = c.AddFunc("0 0 4 * * *", crons.CronRegister[crons.CronDevelopers].Work)
	log.Critical(err)

	// 12.00
	err = c.AddFunc("0 0 12 * * *", crons.CronRegister[crons.CronInstagram].Work)
	log.Critical(err)

	c.Start()

	helpers.KeepAlive()
}
