package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/Jleagle/recaptcha-go"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/queue"
	"github.com/gamedb/website/social"
	"github.com/gamedb/website/web"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

// These are called so everything as access to configs (viper)
func init() {
	configSetup() // First
	log.Init()    // Second
	helpers.InitSteam()
	helpers.InitMemcache()
	db.InitDS()
	queue.Init()
	web.Init()
}

func main() {

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
			err := http.ListenAndServe(":"+viper.GetString("PORT"), nil)
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

func configSetup() {

	// Checks
	if os.Getenv("STEAM_GOOGLE_APPLICATION_CREDENTIALS") == "" {
		panic("can't see environment variables")
	}

	// Google
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", os.Getenv("STEAM_GOOGLE_APPLICATION_CREDENTIALS"))
		log.Err(err)
	}

	// Recaptcha
	recaptcha.SetSecret(os.Getenv("STEAM_RECAPTCHA_PRIVATE"))

	//
	viper.AutomaticEnv()
	viper.SetEnvPrefix("STEAM")

	// Rabbit
	viper.SetDefault("RABBIT_USER", "guest")
	viper.SetDefault("RABBIT_PASS", "guest")

	// Other
	viper.SetDefault("PORT", "8081")
	viper.SetDefault("ENV", "local")
	viper.SetDefault("MEMCACHE_DSN", "memcache:11211")
	viper.SetDefault("PATH", "/root")
	viper.SetDefault("MYSQL_DSN", "root@tcp(localhost:3306)/steam")
	viper.SetDefault("DOMAIN", "https://gamedb.online")
	viper.SetDefault("SHORT_NAME", "GameDB")
}
