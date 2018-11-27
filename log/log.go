package log

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"cloud.google.com/go/logging"
	"github.com/rollbar/rollbar-go"
	"github.com/spf13/viper"
)

type Severity = logging.Severity
type LogName string
type Environment string
type Service string

//noinspection GoUnusedGlobalVariable
var (
	env Environment

	// Google
	googleCtx    = context.Background()
	googleClient *logging.Client

	// Environments
	EnvProd  Environment = "production"
	EnvLocal Environment = "local"

	// Log names
	LogConsumers LogName = "gamedb.consumers"
	LogSteam     LogName = "gamedb.steam"
	LogGameDB    LogName = "gamedb"

	// Severities
	Default   = logging.Default
	Debug     = logging.Debug
	Info      = logging.Info
	Notice    = logging.Notice
	Warning   = logging.Warning
	Error     = logging.Error
	Critical  = logging.Critical
	Alert     = logging.Alert
	Emergency = logging.Emergency

	// Services
	ServiceGoogle  Service = "google"
	ServiceRollbar Service = "rollbar"
	ServiceLocal   Service = "local"
)

// Called from main
func Init() {

	envString := viper.GetString("ENV")
	env = Environment(envString)

	// Setup Google
	var err error
	googleClient, err = logging.NewClient(googleCtx, viper.GetString("GOOGLE_PROJECT"))
	fmt.Println(err)

	// Setup Roolbar
	rollbar.SetToken(viper.GetString("ROLLBAR_PRIVATE"))
	rollbar.SetEnvironment(envString)                  // defaults to "development"
	rollbar.SetCodeVersion("master")                   // optional Git hash/branch/tag (required for GitHub integration)
	rollbar.SetServerRoot("github.com/gamedb/website") // path of project (required for GitHub integration and non-project stacktrace collapsing)
}

func log(interfaces ...interface{}) {

	var services = []Service{ServiceGoogle}
	var logs []*logging.Logger
	var entry logging.Entry

	for _, v := range interfaces {

		if v == nil {
			continue
		}

		switch val := v.(type) {
		case string:
			entry.Payload = val
		case *http.Request:
			entry.HTTPRequest = &logging.HTTPRequest{Request: val}
		case error:
			entry.Payload = val.Error()
		case LogName:
			logs = append(logs, googleClient.Logger(string(val)+"-"+string(env)))
		case Severity:
			entry.Severity = val
		case time.Time:
			entry.Timestamp = val
		case Service:
			services = append(services, val)
		default:
			Text("Invalid value given to Err")
		}
	}

	// Default log
	if len(logs) == 0 {
		logs = append(logs, googleClient.Logger(string(LogGameDB)+"-"+string(env)))
	}

	// Add stack to payload
	entry.Payload = entry.Payload.(string) + "\n\r" + string(debug.Stack())

	for _, log := range logs {
		log.Log(entry)
	}
}

func Err(err error, interfaces ...interface{}) {
	log(append(interfaces, err)...)
}

func Text(text string, interfaces ...interface{}) {
	log(append(interfaces, text)...)
}
