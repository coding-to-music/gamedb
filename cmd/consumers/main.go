package main

import (
	"github.com/gamedb/website/pkg/helpers"
	"github.com/gamedb/website/pkg/log"
	"github.com/gamedb/website/pkg/queue"
)

func main() {

	log.Info("Starting consumers")

	for queueName, q := range queue.QueueRegister {
		q.Name = queueName
		go q.ConsumeMessages()
	}

	helpers.KeepAlive()
}
