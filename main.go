package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"

	"github.com/Jleagle/recaptcha-go"
	"github.com/rollbar/rollbar-go"
	"github.com/spf13/viper"
	"github.com/steam-authority/steam-authority/config"
	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/queue"
	"github.com/steam-authority/steam-authority/session"
	"github.com/steam-authority/steam-authority/storage"
	"github.com/steam-authority/steam-authority/web"
)

// These are called so everything as access to configs (viper)
func init() {
	config.Init() // Must go first
	queue.Init()
	logger.Init()
	session.Init()
	storage.Init()
	web.InitChat()
	web.InitCommits()
}

func main() {

	// Rollbar
	rollbar.SetToken(viper.GetString("ROLLBAR_PRIVATE"))
	rollbar.SetEnvironment(viper.GetString("ENV"))                      // defaults to "development"
	rollbar.SetCodeVersion("master")                                    // optional Git hash/branch/tag (required for GitHub integration)
	rollbar.SetServerRoot("github.com/steam-authority/steam-authority") // path of project (required for GitHub integration and non-project stacktrace collapsing)

	// Recaptcha
	recaptcha.SetSecret(viper.GetString("RECAPTCHA_PRIVATE"))

	// Flags
	flagPprof := flag.Bool("pprof", false, "PProf")
	flagDebug := flag.Bool("debug", false, "Debug")
	flagConsumers := flag.Bool("consumers", true, "Consumers")

	flag.Parse()

	if *flagPprof {
		go http.ListenAndServe(":"+viper.GetString("PORT"), nil)
	}

	if *flagDebug {
		db.SetDebug(true)
	}

	if *flagConsumers {
		queue.RunConsumers()
	}

	// Log steam calls
	go func() {
		for v := range helpers.GetSteamLogsChan() {
			logger.Info(v)
			logger.InfoG(v)
		}
	}()

	// Web server
	err := web.Serve()
	if err != nil {

		logger.Error(err)

	} else {

		// Block forever for goroutines to run
		forever := make(chan bool)
		<-forever
	}
}
