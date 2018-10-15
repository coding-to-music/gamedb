package logger

import (
	"log"
	"os"

	"github.com/rollbar/rollbar-go"
	"github.com/spf13/viper"
)

const (
	EnvProd  = "production"
	EnvLocal = "local"
)

var (
	logger = log.New(os.Stderr, "gamedb: ", log.LstdFlags|log.Lshortfile)
)

func Error(err error) {

	if err != nil {

		logger.Println(err.Error())

		if viper.GetString("ENV") == EnvProd {
			rollbar.Error(rollbar.ERR, err)
			ErrorG(err)
		}
	}
}

func Info(message string) {

	if message != "" {

		logger.Println(message)

		if viper.GetString("ENV") == EnvProd {
			rollbar.Message(rollbar.INFO, message)
			InfoG(message)
		}
	}
}

func LocalInfo(message string) {

	if message != "" && viper.GetString("ENV") == EnvLocal {
		logger.Println(message)
	}
}
