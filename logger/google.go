package logger

import (
	"context"
	"runtime/debug"

	"cloud.google.com/go/logging"
	"github.com/spf13/viper"
)

const (
	LogConsumers = "gamedb.consumers"
	LogGameDB    = "gamedb"
)

var (
	ctx    = context.Background()
	client *logging.Client
)

// Called from main
func Init() {
	var err error
	client, err = logging.NewClient(ctx, viper.GetString("GOOGLE_PROJECT"))
	Error(err)
}

func getLog(name ...string) (*logging.Logger) {

	env := viper.GetString("ENV")

	if len(name) > 0 {
		return client.Logger(name[0] + "-" + env)
	} else {
		return client.Logger(LogGameDB + "-" + env)
	}
}

func ErrorG(err error, log ...string) {
	getLog(log...).Log(logging.Entry{Severity: logging.Error, Payload: err.Error() + "\n\r" + string(debug.Stack())})
}

func InfoG(payload string, log ...string) {
	getLog(log...).Log(logging.Entry{Severity: logging.Info, Payload: payload + "\n\r" + string(debug.Stack())})
}

func CriticalG(err error, log ...string) {
	getLog(log...).LogSync(ctx, logging.Entry{Severity: logging.Critical, Payload: err.Error() + "\n\r" + string(debug.Stack())})
}
