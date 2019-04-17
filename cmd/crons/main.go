package main

import (
	"time"

	"github.com/gamedb/website/cmd/web_server/pages"
	"github.com/gamedb/website/pkg"
	"github.com/robfig/cron"
)

func main() {

	var err error

	if config.Config.IsProd() {

		c := cron.New()

		// Daily
		err = c.AddFunc("1 0 0 * * *", pages.ClearUpcomingCache)
		pkg.Critical(err)

		err = c.AddFunc("0 0 0 * * *", pages.CronRanks)
		pkg.Critical(err)

		err = c.AddFunc("0 0 1 * * *", pages.CronGenres)
		pkg.Critical(err)

		err = c.AddFunc("0 0 2 * * *", pages.CronTags)
		pkg.Critical(err)

		err = c.AddFunc("0 0 3 * * *", pages.CronPublishers)
		pkg.Critical(err)

		err = c.AddFunc("0 0 4 * * *", pages.CronDevelopers)
		pkg.Critical(err)

		err = c.AddFunc("0 0 5 * * *", pages.CronDonations)
		pkg.Critical(err)

		err = c.AddFunc("0 0 12 * * *", pkg.UploadInstagram)
		pkg.Critical(err)

		// Every 3 hours
		err = c.AddFunc("0 0 */3 * * *", pages.CronCheckForPlayers)
		pkg.Critical(err)

		c.Start()

		// Scan for app players after deploy
		go func() {
			time.Sleep(time.Minute)
			pages.CronCheckForPlayers()
		}()
	}

	pkg.KeepAlive()
}
