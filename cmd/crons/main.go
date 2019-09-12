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
	log.SetVersion(version)

	log.Info("Starting crons")

	rand.Seed(time.Now().Unix())

	var err error

	c := cron.New(
		cron.WithLogger(cronLogger{}),
		cron.WithParser(cron.NewParser(cron.Minute|cron.Hour)),
	)

	for _, v := range tasks.TaskRegister {
		if v.Cron() != "" {
			log.Info("Adding " + v.ID())
			_, err = c.AddFunc(v.Cron(), func() { tasks.RunTask(v) })
			log.Critical(err)
		}
	}

	c.Start()

	helpers.KeepAlive()
}

type cronLogger struct {
}

func (cl cronLogger) Info(msg string, keysAndValues ...interface{}) {
	// log.Info(msg)
}

func (cl cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	log.Err(msg, err)
}
