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
		return client.Logger(env + "_" + name[0])
	} else {
		return client.Logger(env + "_" + LogGameDB)
	}
}

func ErrorG(err error, log ...string) {
	getLog(log...).Log(logging.Entry{Payload: err.Error() + "\n\r" + string(debug.Stack()), Severity: logging.Error})
}

func InfoG(payload string, log ...string) {
	getLog(log...).Log(logging.Entry{Payload: payload, Severity: logging.Info})
}

func CriticalG(err error, log ...string) {
	getLog(log...).LogSync(ctx, logging.Entry{Payload: err.Error(), Severity: logging.Critical})
}
