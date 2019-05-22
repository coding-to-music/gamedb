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
	err = c.AddFunc("0 */10 * * * *", crons.CronRegister[crons.CronSteamClientPlayers])
	log.Critical(err)

	// Every 6 hours
	err = c.AddFunc("0 0 */6 * * *", crons.CronRegister[crons.CronAppPlayers])
	log.Critical(err)

	err = c.AddFunc("0 0 */6 * * *", crons.CronRegister[crons.CronAutoPlayerRefreshes])
	log.Critical(err)

	// Every 24 hours
	err = c.AddFunc("0 1 0 * * *", crons.CronRegister[crons.CronClearUpcomingCache])
	log.Critical(err)

	err = c.AddFunc("0 0 0 * * *", crons.CronRegister[crons.CronPlayerRanks])
	log.Critical(err)

	err = c.AddFunc("0 0 1 * * *", crons.CronRegister[crons.CronGenres])
	log.Critical(err)

	err = c.AddFunc("0 0 2 * * *", crons.CronRegister[crons.CronTags])
	log.Critical(err)

	err = c.AddFunc("0 0 3 * * *", crons.CronRegister[crons.CronPublishers])
	log.Critical(err)

	err = c.AddFunc("0 0 4 * * *", crons.CronRegister[crons.CronDevelopers])
	log.Critical(err)

	err = c.AddFunc("0 0 12 * * *", crons.CronRegister[crons.CronInstagram])
	log.Critical(err)

	c.Start()

	// // Scan for app players after deploy
	// go func() {
	// 	time.Sleep(time.Minute)
	// 	crons.AppPlayers()
	// }()

	helpers.KeepAlive()
}
