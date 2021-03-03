package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/crons"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/robfig/cron/v3"
)

func main() {

	err := config.Init(helpers.GetIP())
	log.InitZap(log.LogNameCrons)
	defer log.Flush()
	if err != nil {
		log.ErrS(err)
		return
	}

	// Load queue producers
	queue.Init(queue.QueueCronsDefinitions)

	// Profiling
	go func() {
		err := http.ListenAndServe(":6063", nil)
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get API key
	err = mysql.GetConsumer("crons")
	if err != nil {
		log.ErrS(err)
		return
	}

	c := cron.New(
		cron.WithLogger(cronLogger{}),
		cron.WithParser(crons.Parser),
	)

	for _, task := range crons.TaskRegister {
		// In a func here so `task` gets copied into a new memory location and can not be replaced at a later time
		func(task crons.TaskInterface) {
			if task.Cron() != "" {
				_, err := c.AddFunc(string(task.Cron()), func() { crons.Run(task) })
				if err != nil {
					log.ErrS(err, task.ID())
				}
			}
		}(task)
	}

	log.Info("Starting crons")
	go c.Run() // Blocks

	helpers.KeepAlive(
		mysql.Close,
		mongo.Close,
	)
}

//
type cronLogger struct {
}

func (cl cronLogger) Info(msg string, keysAndValues ...interface{}) {

	// is := []interface{}{msg}
	// is = append(is, keysAndValues...)

	// log.ErrS(is...)
}

func (cl cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {

	is := []interface{}{msg, err}
	is = append(is, keysAndValues...)

	log.ErrS(is...)
}
