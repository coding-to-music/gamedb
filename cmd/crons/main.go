package main

import (
	"math/rand"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/tasks"
	"github.com/robfig/cron"
)

func main() {

	log.Info("Starting crons")

	rand.Seed(time.Now().Unix())

	var err error

	c := cron.New()

	for _, v := range tasks.TaskRegister {
		if v.Cron() != "" {
			log.Info(v.Name())
			err = c.AddFunc(v.Cron(), func() { tasks.RunTask(v) })
			log.Critical(err)
		}
	}

	c.Start()

	helpers.KeepAlive()
}
