package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/rollbar/rollbar-go"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/pics"
	"github.com/steam-authority/steam-authority/queue"
	"github.com/steam-authority/steam-authority/web"
)

func main() {

	// Rollbar
	rollbar.SetToken(os.Getenv("STEAM_ROLLBAR_PRIVATE"))
	rollbar.SetEnvironment(os.Getenv("ENV"))                            // defaults to "development"
	rollbar.SetCodeVersion("dev-master")                                // optional Git hash/branch/tag (required for GitHub integration)
	rollbar.SetServerRoot("github.com/steam-authority/steam-authority") // path of project (required for GitHub integration and non-project stacktrace collapsing)

	// Env vars
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", os.Getenv("STEAM_GOOGLE_APPLICATION_CREDENTIALS"))
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
		go pics.Run()
	}

	if *flagConsumers {
		queue.RunConsumers()
	}

	// Web server
	web.Serve()

	// Block for goroutines to run forever
	forever := make(chan bool)
	<-forever
}
