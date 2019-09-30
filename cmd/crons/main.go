package main

import (
	"math/rand"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/tasks"
	"github.com/robfig/cron/v3"
)

var version string

func main() {

	config.SetVersion(version)
	log.Initialise()

	log.Info("Starting crons")

	rand.Seed(time.Now().Unix())

	c := cron.New(
		cron.WithLogger(cronLogger{}),
		cron.WithParser(tasks.Parser),
		cron.WithChain(
			cron.SkipIfStillRunning(cronLogger{}),
		),
	)

	c.AddFunc(tasks.AppPlayers{}.Cron(), func() { tasks.AppPlayers{}.Run() })
	c.AddFunc(tasks.AutoPlayerRefreshes{}.Cron(), func() { tasks.AutoPlayerRefreshes{}.Run() })
	c.AddFunc(tasks.ClearUpcomingCache{}.Cron(), func() { tasks.ClearUpcomingCache{}.Run() })
	c.AddFunc(tasks.Developers{}.Cron(), func() { tasks.Developers{}.Run() })
	c.AddFunc(tasks.Genres{}.Cron(), func() { tasks.Genres{}.Run() })
	c.AddFunc(tasks.Instagram{}.Cron(), func() { tasks.Instagram{}.Run() })
	c.AddFunc(tasks.SetBadgeCache{}.Cron(), func() { tasks.SetBadgeCache{}.Run() })
	c.AddFunc(tasks.PlayerRanks{}.Cron(), func() { tasks.PlayerRanks{}.Run() })
	c.AddFunc(tasks.Publishers{}.Cron(), func() { tasks.Publishers{}.Run() })
	c.AddFunc(tasks.SteamClientPlayers{}.Cron(), func() { tasks.SteamClientPlayers{}.Run() })
	c.AddFunc(tasks.Tags{}.Cron(), func() { tasks.Tags{}.Run() })
	c.AddFunc(tasks.Wishlists{}.Cron(), func() { tasks.Wishlists{}.Run() })

	c.Start()

	helpers.KeepAlive()
}

type cronLogger struct {
}

func (cl cronLogger) Info(msg string, keysAndValues ...interface{}) {
	log.Info(msg)
}

func (cl cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	log.Err(msg, err)
}
