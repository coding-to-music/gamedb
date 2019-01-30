package main

import (
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gamedb/website/config"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/queue"
	"github.com/gamedb/website/web"
	_ "github.com/go-sql-driver/mysql"
)

func main() {

	rand.Seed(time.Now().UnixNano())

	// Try to stop memcache failing on first calls..
	helpers.GetMemcache()

	// Web server
	if config.Config.EnableWebserver.GetBool() {
		go func() {
			log.Info("Starting web server")
			err := web.Serve()
			log.Err(err)
		}()
	}

	// Consumers
	if config.Config.EnableConsumers.GetBool() {
		go func() {
			log.Info("Starting consumers")
			queue.RunConsumers()
		}()
	}

	// Log steam calls
	go func() {
		for v := range helpers.GetSteamLogsChan() {
			log.Info(log.ServiceGoogle, v.String(), log.LogNameSteam)
		}
	}()

	// Log number of goroutines
	go func() {
		for {
			time.Sleep(time.Minute * 10)
			log.Info("Goroutines running: "+strconv.Itoa(runtime.NumGoroutine()), log.SeverityInfo, log.ServiceGoogle)
		}
	}()

	// Instagram
	if config.Config.IsProd() {
		go func() {
			helpers.RunInstagram()
		}()
	}

	// Block forever for goroutines to run
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
