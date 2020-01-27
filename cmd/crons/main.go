package main

import (
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/tasks"
	"github.com/robfig/cron/v3"
)

var version string

func main() {

	config.SetVersion(version)
	log.Initialise([]log.LogName{log.LogNameCrons})

	// Load queue producers
	queue.Init(queue.QueueCronsDefinitions, false)

	// Get API key
	err := sql.GetAPIKey("crons")
	if err != nil {
		log.Critical(err)
		return
	}

	c := cron.New(
		cron.WithLogger(cronLogger{}),
		cron.WithParser(tasks.Parser),
	)

	for _, task := range tasks.TaskRegister {
		// In a func here so `task` gets copied into a new memory location and can not be replaced at a later time
		func(task tasks.TaskInterface) {
			if task.Cron() != "" {
				_, err := c.AddFunc(task.Cron(), func() { tasks.Run(task) })
				log.Err(err)
			}
		}(task)
	}

	log.Info("Starting crons")
	c.Run() // Blocks
}

//
type cronLogger struct {
}

func (cl cronLogger) Info(msg string, keysAndValues ...interface{}) {
	// log.Info(msg)
}

func (cl cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	log.Err(msg, err)
}
