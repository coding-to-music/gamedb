package logging

import (
	"github.com/rollbar/rollbar-go"
	"github.com/spf13/viper"
)

func ErrorR(err error, local ...bool) {

	if len(local) > 0 {
		ErrorL(err)
	}

	if err != nil {
		if viper.GetString("ENV") == EnvProd {
			rollbar.Error(rollbar.ERR, err)
		}
	}
}

func InfoR(message string, local ...bool) {

	if len(local) > 0 {
		InfoL(message)
	}

	if message != "" {
		if viper.GetString("ENV") == EnvProd {
			rollbar.Message(rollbar.INFO, message)
		}
	}
}
