package main

import (
	"math/rand"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/tasks"
	"github.com/robfig/cron"
)

var version string

func main() {

	log.Info("Starting crons")

	rand.Seed(time.Now().Unix())

	config.Config.CommitHash.SetDefault(version)

	var err error

	c := cron.New()

	for _, v := range tasks.TaskRegister {
		if v.Cron() != "" {
			log.Info("Adding " + v.Name() + " to cron")
			err = c.AddFunc(v.Cron(), func() { tasks.RunTask(v) })
			log.Critical(err)
		}
	}

	c.Start()

	helpers.KeepAlive()
}
