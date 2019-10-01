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
	log.Initialise([]log.LogName{log.LogNameCrons})

	log.Info("Starting crons")

	rand.Seed(time.Now().Unix())

	c := cron.New(
		cron.WithLogger(cronLogger{}),
		cron.WithParser(tasks.Parser),
		cron.WithChain(
			cron.SkipIfStillRunning(cronLogger{}),
		),
	)

	c.AddFunc(tasks.AppPlayers{}.Cron(), func() { tasks.TaskRegister[tasks.AppPlayers{}.ID()].Run() })
	c.AddFunc(tasks.AutoPlayerRefreshes{}.Cron(), func() { tasks.TaskRegister[tasks.AutoPlayerRefreshes{}.ID()].Run() })
	c.AddFunc(tasks.ClearUpcomingCache{}.Cron(), func() { tasks.TaskRegister[tasks.ClearUpcomingCache{}.ID()].Run() })
	c.AddFunc(tasks.Developers{}.Cron(), func() { tasks.TaskRegister[tasks.Developers{}.ID()].Run() })
	c.AddFunc(tasks.Genres{}.Cron(), func() { tasks.TaskRegister[tasks.Genres{}.ID()].Run() })
	c.AddFunc(tasks.Instagram{}.Cron(), func() { tasks.TaskRegister[tasks.Instagram{}.ID()].Run() })
	c.AddFunc(tasks.SetBadgeCache{}.Cron(), func() { tasks.TaskRegister[tasks.SetBadgeCache{}.ID()].Run() })
	c.AddFunc(tasks.PlayerRanks{}.Cron(), func() { tasks.TaskRegister[tasks.PlayerRanks{}.ID()].Run() })
	c.AddFunc(tasks.Publishers{}.Cron(), func() { tasks.TaskRegister[tasks.Publishers{}.ID()].Run() })
	c.AddFunc(tasks.SteamClientPlayers{}.Cron(), func() { tasks.TaskRegister[tasks.SteamClientPlayers{}.ID()].Run() })
	c.AddFunc(tasks.Tags{}.Cron(), func() { tasks.TaskRegister[tasks.Tags{}.ID()].Run() })
	c.AddFunc(tasks.Wishlists{}.Cron(), func() { tasks.TaskRegister[tasks.Wishlists{}.ID()].Run() })

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
