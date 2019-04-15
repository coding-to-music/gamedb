package main

import (
	"math/rand"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Jleagle/recaptcha-go"
	"github.com/gamedb/website/chat_bot"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/queue"
	"github.com/gamedb/website/social"
	"github.com/gamedb/website/sql"
	"github.com/gamedb/website/web"
	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/cron"
)

//noinspection GoUnusedGlobalVariable
var version string

func main() {

	config.Config.CommitHash = version

	recaptcha.SetSecret(config.Config.RecaptchaPrivate)

	rand.Seed(time.Now().UnixNano())

	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		log.Err("GOOGLE_APPLICATION_CREDENTIALS not found")
		os.Exit(1)
	}

	// Preload connections
	helpers.GetMemcache()
	_, err := sql.GetMySQLClient()
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

	// Chat Bot
	if config.Config.IsProd() || config.Config.IsLocal() {
		chat_bot.Init()
	}

	// Crons
	if config.Config.IsProd() {

		c := cron.New()

		// Daily
		err = c.AddFunc("1 0 0 * * *", web.ClearUpcomingCache)
		log.Critical(err)

		err = c.AddFunc("0 0 0 * * *", web.CronRanks)
		log.Critical(err)

		err = c.AddFunc("0 0 1 * * *", web.CronGenres)
		log.Critical(err)

		err = c.AddFunc("0 0 2 * * *", web.CronTags)
		log.Critical(err)

		err = c.AddFunc("0 0 3 * * *", web.CronPublishers)
		log.Critical(err)

		err = c.AddFunc("0 0 4 * * *", web.CronDevelopers)
		log.Critical(err)

		err = c.AddFunc("0 0 5 * * *", web.CronDonations)
		log.Critical(err)

		err = c.AddFunc("0 0 12 * * *", social.UploadInstagram)
		log.Critical(err)

		// Every 3 hours
		err = c.AddFunc("0 0 */3 * * *", web.CronCheckForPlayers)
		log.Critical(err)

		c.Start()

		// Scan for app players after deploy
		go func() {
			time.Sleep(time.Minute)
			web.CronCheckForPlayers()
		}()
	}

	// Block forever for goroutines to run
	x := make(chan os.Signal)
	signal.Notify(x, syscall.SIGTERM, os.Interrupt, os.Kill)

	wg := &sync.WaitGroup{} // Must be pointer
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		for range x {

			//noinspection GoDeferInLoop
			defer wg.Done()

			client, err := sql.GetMySQLClient()
			if err != nil {
				log.Err(err)
				return
			}

			err = client.Close()
			log.Err(err)

			return
		}
	}(wg)

	wg.Wait()
}
