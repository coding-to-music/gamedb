package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/Jleagle/recaptcha-go"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/logging"
	"github.com/gamedb/website/queue"
	"github.com/gamedb/website/storage"
	"github.com/gamedb/website/web"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

// These are called so everything as access to configs (viper)
func init() {
	configSetup()  // First
	logging.Init() // Second
	log.Init()     // Second
	helpers.InitSteam()
	helpers.InitMemcache()
	db.InitDS()
	storage.Init()
	queue.Init()
	web.Init()
}

func main() {

	// Recaptcha
	recaptcha.SetSecret(viper.GetString("RECAPTCHA_PRIVATE"))

	// Flags
	flagPprof := flag.Bool("pprof", false, "PProf")
	flagConsumers := flag.Bool("consumers", true, "Consumers")

	flag.Parse()

	if *flagPprof {
		go func() {
			err := http.ListenAndServe(":"+viper.GetString("PORT"), nil)
			logging.Error(err)
		}()
	}

	if *flagConsumers {
		queue.RunConsumers()
	}

	// Log steam calls
	go func() {
		for v := range helpers.GetSteamLogsChan() {
			logging.InfoG(v.String(), logging.LogSteam)
		}
	}()

	// Web server
	err := web.Serve()
	if err != nil {

		logging.Error(err)

	} else {

		// Block forever for goroutines to run
		forever := make(chan bool)
		<-forever
	}
}

func configSetup() {

	// Checks
	if os.Getenv("STEAM_GOOGLE_APPLICATION_CREDENTIALS") == "" {
		panic("can't see environment variables")
	}

	// Google
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", os.Getenv("STEAM_GOOGLE_APPLICATION_CREDENTIALS"))
		logging.Error(err)
	}

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
