package logger

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/logging"
)

const (
	LogPics      = "pics"
	LogConsumers = "consumers"
)

var (
	ctx    = context.Background()
	client *logging.Client
)

func init() {
	var err error
	client, err = logging.NewClient(ctx, os.Getenv("STEAM_GOOGLE_PROJECT"))
	if err != nil {
		fmt.Println(err.Error())
	}
}

func getLog(name ...string) (*logging.Logger) {

	env := os.Getenv("ENV")

	if len(name) > 0 {
		return client.Logger(env + "_" + name[0])
	} else {
		return client.Logger(env + "_" + "main")
	}
}

//func Error(err error, log ...string) {
//	getLog(log...).Log(logging.Entry{Payload: err.Error() + "\n\r" + string(debug.Stack()), Severity: logging.Error})
//}
//
//func Info(payload string, log ...string) {
//	getLog(log...).Log(logging.Entry{Payload: payload, Severity: logging.Info})
//}
//
//func Critical(err error, log ...string) {
//	getLog(log...).LogSync(ctx, logging.Entry{Payload: err.Error(), Severity: logging.Critical})
//}
