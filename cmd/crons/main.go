package main

import (
	"math/rand"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
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

	for _, task := range tasks.TaskRegister {
		go func(task tasks.TaskInterface) {
			task = tasks.TaskRegister[task.ID()]
			if task.Cron() != "" {
				_, err := c.AddFunc(task.Cron(), func() { tasks.Run(task) })
				log.Err(err)
			}
		}(task)
	}

	c.Run() // Blocks
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
