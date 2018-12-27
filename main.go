package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/gamedb/website/config"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/queue"
	"github.com/gamedb/website/social"
	"github.com/gamedb/website/web"
	_ "github.com/go-sql-driver/mysql"
)

func main() {

	log.Info("Main: " + config.Config.Environment.Get())

	// Flags
	flagWebServer := flag.Bool("webserver", false, "Web Server")
	flagConsumers := flag.Bool("consumers", false, "Consumers")
	flagPprof := flag.Bool("pprof", false, "PProf")
	instagram := flag.Bool("instagram", false, "Instagram")

	flag.Parse()

	if *instagram {
		social.InitIG()
		os.Exit(0)
	}

	// Web server
	if *flagWebServer {
		go func() {
			log.Info("Starting web server")
			err := web.Serve()
			log.Err(err)
		}()
	}

	if *flagConsumers {
		go func() {
			log.Info("Starting consumers")
			queue.RunConsumers()
		}()
	}

	if *flagPprof {
		go func() {
			log.Info("Starting pprof")
			err := http.ListenAndServe(config.Config.ListenOn(), nil)
			log.Err(err)
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
			log.Info("Goroutines running: "+strconv.Itoa(runtime.NumGoroutine()), log.SeverityInfo)
		}
	}()

	// Block forever for goroutines to run
	forever := make(chan bool)
	<-forever
}
