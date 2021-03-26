package main

import (
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
)

func main() {

	rand.Seed(time.Now().UnixNano())

	err := config.Init(helpers.GetIP())
	log.InitZap(log.LogNameConsumers)
	defer log.Flush()
	if err != nil {
		log.ErrS(err)
		return
	}

	//
	log.Info("Starting consumers")

	// Get API key
	err = mysql.GetConsumer("consumer")
	if err != nil {
		log.ErrS(err)
		return
	}

	// Profiling
	if config.IsLocal() {
		go func() {
			err := http.ListenAndServe(":6062", nil)
			if err != nil {
				log.ErrS(err)
			}
		}()
	}

	// Load consumers
	consumers.Init(consumers.ConsumersDefinitions)

	helpers.KeepAlive(
		mysql.Close,
		mongo.Close,
		memcache.Close,
	)
}
