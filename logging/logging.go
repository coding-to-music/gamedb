package logging

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/logging"
	"github.com/spf13/viper"
)

const (
	EnvProd  = "production"
	EnvLocal = "local"

	LogConsumers = "gamedb.consumers"
	LogSteam     = "gamedb.steam"
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

var (
	logger = log.New(os.Stderr, "gamedb: ", log.LstdFlags)
)

func Error(err error) {

	ErrorL(err)
	ErrorG(err)
	//ErrorR(err)
}

func Info(message string) {

	InfoL(message)
	InfoG(message)
	//InfoR(message)
}
