package logging

import (
	"github.com/go-errors/errors"
	"github.com/spf13/viper"
)

func ErrorL(err error) {

	if err != nil && viper.GetString("ENV") == EnvLocal {

		err2 := errors.Wrap(err, 2)

		logger.Println(err2.Error() + " - " + err2.ErrorStack())
	}
}

func InfoL(message string) {

	if message != "" && viper.GetString("ENV") == EnvLocal {
		logger.Println(message)
	}
}
