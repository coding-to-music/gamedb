package main

import (
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gamedb/website/config"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/queue"
	"github.com/gamedb/website/social"
	"github.com/gamedb/website/web"
	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/cron"
)

func main() {

	rand.Seed(time.Now().UnixNano())

	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		log.Err("GOOGLE_APPLICATION_CREDENTIALS not found")
		os.Exit(1)
	}

	// Preload connections
	helpers.GetMemcache()
	_, err := db.GetMySQLClient()
	log.Critical(err)

	// Web server
	if config.Config.EnableWebserver.GetBool() {
		go func() {
			log.Info("Starting web server")
			err := web.Serve()
			log.Critical(err)
		}()
	}

	// Consumers
	if config.Config.EnableConsumers.GetBool() {
		go func() {
			log.Info("Starting consumers")
			queue.RunConsumers()
		}()
	}

	// Log number of goroutines
	go func() {
		log.Info("Logging goroutines")
		for {
			time.Sleep(time.Minute * 10)
			log.Info("Goroutines running: "+strconv.Itoa(runtime.NumGoroutine()), log.SeverityInfo, log.ServiceGoogle)
		}
	}()

	// Crons
	if config.Config.IsProd() {
		c := cron.New()

		err = c.AddFunc("0 0 0 * * *", web.CronRanks)
		log.Critical(err)

		// err = c.AddFunc("0 0 1 * * *", web.CronGenres)
		// log.Critical(err)

		// err = c.AddFunc("0 0 2 * * *", web.CronTags)
		// log.Critical(err)

		// err = c.AddFunc("0 0 3 * * *", web.CronPublishers)
		// log.Critical(err)

		// err = c.AddFunc("0 0 4 * * *", web.CronDevelopers)
		// log.Critical(err)

		err = c.AddFunc("0 0 5 * * *", web.CronDonations)
		log.Critical(err)

		err = c.AddFunc("0 0 12 * * *", social.UploadInstagram)
		log.Critical(err)

		c.Start()
	}

	// Block forever for goroutines to run
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, os.Interrupt, os.Kill)

	wg := &sync.WaitGroup{} // Must be pointer
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		for range c {

			//noinspection GoDeferInLoop
			defer wg.Done()

			sql, err := db.GetMySQLClient()
			log.Err(err)

			err = sql.Close()
			log.Err(err)

			return
		}
	}(wg)

	wg.Wait()
}
