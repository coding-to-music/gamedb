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
	)

	c.AddFunc("*/10 *", func() { tasks.Run(tasks.TaskRegister[tasks.SetBadgeCache{}.ID()]) })
	c.AddFunc("*/10 *", func() { tasks.Run(tasks.TaskRegister[tasks.SteamClientPlayers{}.ID()]) })

	c.AddFunc("0 0", func() { tasks.Run(tasks.TaskRegister[tasks.Wishlists{}.ID()]) })
	c.AddFunc("1 0", func() { tasks.Run(tasks.TaskRegister[tasks.ClearUpcomingCache{}.ID()]) })
	c.AddFunc("2 0", func() { tasks.Run(tasks.TaskRegister[tasks.PlayerRanks{}.ID()]) })
	c.AddFunc("0 1", func() { tasks.Run(tasks.TaskRegister[tasks.Genres{}.ID()]) })
	c.AddFunc("0 2", func() { tasks.Run(tasks.TaskRegister[tasks.Tags{}.ID()]) })
	c.AddFunc("0 3", func() { tasks.Run(tasks.TaskRegister[tasks.Publishers{}.ID()]) })
	c.AddFunc("0 4", func() { tasks.Run(tasks.TaskRegister[tasks.Developers{}.ID()]) })
	c.AddFunc("0 12", func() { tasks.Run(tasks.TaskRegister[tasks.Instagram{}.ID()]) })

	c.AddFunc("0 */5", func() { tasks.Run(tasks.TaskRegister[tasks.AppPlayers{}.ID()]) })
	c.AddFunc("0 */6", func() { tasks.Run(tasks.TaskRegister[tasks.AutoPlayerRefreshes{}.ID()]) })

	c.Start()

	helpers.KeepAlive()
}

//
type cronLogger struct {
}

func (cl cronLogger) Info(msg string, keysAndValues ...interface{}) {
	log.Info(msg)
}

func (cl cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	log.Err(msg, err)
}
