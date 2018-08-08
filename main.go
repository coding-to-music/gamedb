package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/Jleagle/recaptcha-go"
	"github.com/rollbar/rollbar-go"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/queue"
	"github.com/steam-authority/steam-authority/web"
)

func main() {

	// Rollbar
	rollbar.SetToken(os.Getenv("STEAM_ROLLBAR_PRIVATE"))
	rollbar.SetEnvironment(os.Getenv("ENV"))                            // defaults to "development"
	rollbar.SetCodeVersion("master")                                    // optional Git hash/branch/tag (required for GitHub integration)
	rollbar.SetServerRoot("github.com/steam-authority/steam-authority") // path of project (required for GitHub integration and non-project stacktrace collapsing)

	// Recaptcha
	recaptcha.SetSecret(os.Getenv("STEAM_RECAPTCHA_PRIVATE"))

	// Google
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", os.Getenv("STEAM_GOOGLE_APPLICATION_CREDENTIALS"))
	}

	// Env vars
	if os.Getenv("ENV") == "local" {
		os.Setenv("STEAM_DOMAIN", os.Getenv("STEAM_DOMAIN_LOCAL"))
	} else {
		os.Setenv("STEAM_DOMAIN", "https://steamauthority.net")
	}

	// Flags
	flagDebug := flag.Bool("debug", false, "Debug")
	flagPics := flag.Bool("pics", false, "Pics")
	flagConsumers := flag.Bool("consumers", false, "Consumers")
	flagPprof := flag.Bool("pprof", false, "PProf")

	flag.Parse()

	if *flagPprof {
		go http.ListenAndServe(":8080", nil)
	}

	if *flagDebug {
		mysql.SetDebug(true)
	}

	if *flagPics {
		//go pics.Run()
	}

	if *flagConsumers {
		queue.RunConsumers()
	}

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
