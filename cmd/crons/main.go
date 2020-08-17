package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/tasks"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

var version string
var commits string

func main() {

	config.Init(version, commits, helpers.GetIP())
	log.InitZap(log.LogNameCrons)

	// Load queue producers
	queue.Init(queue.QueueCronsDefinitions)

	// Profiling
	go func() {
		err := http.ListenAndServe(":6063", nil)
		zap.S().Fatal(err)
	}()

	// Get API key
	err := mysql.GetConsumer("crons")
	if err != nil {
		zap.S().Fatal(err)
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
				_, err := c.AddFunc(string(task.Cron()), func() { tasks.Run(task) })
				if err != nil {
					zap.S().Error(err, task.ID())
				}
			}
		}(task)
	}

	zap.S().Info("Starting crons")
	c.Run() // Blocks
}

//
type cronLogger struct {
}

func (cl cronLogger) Info(msg string, keysAndValues ...interface{}) {

	// is := []interface{}{msg}
	// is = append(is, keysAndValues...)

	// zap.S().Error(is...)
}

func (cl cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {

	is := []interface{}{msg, err}
	is = append(is, keysAndValues...)

	zap.S().Error(is...)
}
