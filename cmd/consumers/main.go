package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
)

var version string

func main() {

	config.SetVersion(version)
	log.Initialise([]log.LogName{log.LogNameConsumers})

	// Get API key
	err := sql.GetAPIKey("consumer", true)
	if err != nil {
		log.Critical(err)
		return
	}

	// Load pubsub
	log.Info("Listening to PubSub for memcache")
	go helpers.ListenToPubSubMemcache()

	// Load PPROF
	if config.IsLocal() {
		log.Info("Starting consumers profiling")
		go func() {
			err := http.ListenAndServe("localhost:6060", nil)
			log.Critical(err)
		}()
	}

	// Load consumers
	log.Info("Starting consumers")
	for queueName, q := range queue.QueueRegister {
		if !q.DoNotScale {
			q.Name = queueName
			go q.ConsumeMessages()
		}
	}

	//
	helpers.KeepAlive()
}
