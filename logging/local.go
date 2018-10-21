package logging

import "github.com/spf13/viper"

func ErrorL(err error) {

	if err != nil && viper.GetString("ENV") == EnvLocal {
		logger.Println(err.Error())
	}
}

func InfoL(message string) {

	if message != "" && viper.GetString("ENV") == EnvLocal {
		logger.Println(message)
	}
}
