package main

import (
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
)

func main() {

	log.Info("Starting consumers")

	for queueName, q := range queue.QueueRegister {
		q.Name = queueName
		go q.ConsumeMessages()
	}

	helpers.KeepAlive()
}
