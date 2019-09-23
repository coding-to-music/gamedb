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

	c.AddFunc(tasks.AppPlayers{}.Cron(), func() { tasks.RunTask(tasks.AppPlayers{}) })
	c.AddFunc(tasks.AutoPlayerRefreshes{}.Cron(), func() { tasks.RunTask(tasks.AutoPlayerRefreshes{}) })
	c.AddFunc(tasks.ClearUpcomingCache{}.Cron(), func() { tasks.RunTask(tasks.ClearUpcomingCache{}) })
	c.AddFunc(tasks.Developers{}.Cron(), func() { tasks.RunTask(tasks.Developers{}) })
	c.AddFunc(tasks.Genres{}.Cron(), func() { tasks.RunTask(tasks.Genres{}) })
	c.AddFunc(tasks.Instagram{}.Cron(), func() { tasks.RunTask(tasks.Instagram{}) })
	c.AddFunc(tasks.SetBadgeCache{}.Cron(), func() { tasks.RunTask(tasks.SetBadgeCache{}) })
	c.AddFunc(tasks.PlayerRanks{}.Cron(), func() { tasks.RunTask(tasks.PlayerRanks{}) })
	c.AddFunc(tasks.Publishers{}.Cron(), func() { tasks.RunTask(tasks.Publishers{}) })
	c.AddFunc(tasks.SteamClientPlayers{}.Cron(), func() { tasks.RunTask(tasks.SteamClientPlayers{}) })
	c.AddFunc(tasks.Tags{}.Cron(), func() { tasks.RunTask(tasks.Tags{}) })
	c.AddFunc(tasks.Wishlists{}.Cron(), func() { tasks.RunTask(tasks.Wishlists{}) })

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
