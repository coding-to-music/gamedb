package logger

import (
	"log"
	"os"

	"github.com/rollbar/rollbar-go"
	"github.com/spf13/viper"
)

const (
	prod  = "production"
	local = "local"
)

var (
	logger = log.New(os.Stderr, "gamedb: ", log.Ltime)
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func Error(err error) {

	if err != nil {

		logger.Println(err.Error())

		if viper.GetString("ENV") == prod {
			rollbar.Error(rollbar.ERR, err)
		}
	}
}

func Info(message string) {

	if message != "" {

		logger.Println(message)

		if viper.GetString("ENV") == prod {
			rollbar.Message(rollbar.INFO, message)
		}
	}
}
