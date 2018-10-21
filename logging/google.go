package logging

import (
	"runtime/debug"

	"cloud.google.com/go/logging"
	"github.com/spf13/viper"
)

func ErrorG(err error, log ...string) {

	if len(log) > 1 {
		ErrorL(err)
	}

	if err != nil {
		getGoogleLog(log...).Log(logging.Entry{Severity: logging.Error, Payload: err.Error() + "\n\r" + string(debug.Stack())})
	}
}

func InfoG(message string, log ...string) {

	if len(log) > 1 {
		InfoL(message)
	}

	if message != "" {
		getGoogleLog(log...).Log(logging.Entry{Severity: logging.Info, Payload: message})
	}
}

func getGoogleLog(name ...string) (*logging.Logger) {

	env := viper.GetString("ENV")

	if len(name) > 0 {
		return client.Logger(name[0] + "-" + env)
	} else {
		return client.Logger(LogGameDB + "-" + env)
	}
}
