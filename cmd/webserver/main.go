package main

import (
	"math/rand"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/Jleagle/recaptcha-go"
	"github.com/gamedb/website/cmd/webserver/pages"
	"github.com/gamedb/website/pkg/config"
	"github.com/gamedb/website/pkg/helpers"
	"github.com/gamedb/website/pkg/log"
	"github.com/gamedb/website/pkg/sql"
	_ "github.com/go-sql-driver/mysql"
)

//noinspection GoUnusedGlobalVariable
var version string

func main() {

	config.Config.CommitHash.SetDefault(version)

	recaptcha.SetSecret(config.Config.RecaptchaPrivate.Get())

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
			err := pages.Serve()
			log.Critical(err)
		}()
	}

	helpers.KeepAlive()
}
