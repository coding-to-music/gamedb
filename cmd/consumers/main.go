package main

import (
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/consumers/framework"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
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
	go memcache.ListenToPubSubMemcache()

	// Load PPROF
	if config.IsLocal() {
		log.Info("Starting consumers profiling")
		go func() {
			err := http.ListenAndServe("localhost:6060", nil)
			log.Critical(err)
		}()
	}

	// Load consumers
	consumers.Init(consumers.QueueDefinitions, true)

	log.Info("Starting old consumers")
	for queueName, q := range queue.QueueRegister {
		if !q.DoNotScale {
			q.Name = queueName
			go q.ConsumeMessages()
		}
	}

	if config.IsProd() {
		go func() {
			for {
				q, err := consumers.Channels[framework.Producer][consumers.QueuePlayers].Inspect()
				if err != nil {
					log.Err(err)
				} else if q.Messages < 10 {
					players, err := mongo.GetRandomPlayers(10)
					if err != nil {
						log.Err(err)
					} else {
						for _, v := range players {
							err = consumers.ProducePlayer(v.ID)
							log.Err(err)
						}
					}

				}

				time.Sleep(time.Second * 10)
			}
		}()
	}

	select {}
}
