package log

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
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

	// Local
	logger = log.New(os.Stderr, "", log.Ltime)

	// Environments
	EnvProd  Environment = "production"
	EnvLocal Environment = "local"

	// Log names
	LogNameConsumers LogName = "gamedb.consumers"
	LogNameSteam     LogName = "gamedb.steam"
	LogNameGameDB    LogName = "gamedb"

	// Severities
	SeverityDefault   = logging.Default
	SeverityDebug     = logging.Debug
	SeverityInfo      = logging.Info
	SeverityNotice    = logging.Notice
	SeverityWarning   = logging.Warning
	SeverityError     = logging.Error
	SeverityCritical  = logging.Critical
	SeverityAlert     = logging.Alert
	SeverityEmergency = logging.Emergency

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
	if err != nil {
		fmt.Println(err)
	}

	// Setup Roolbar
	rollbar.SetToken(viper.GetString("ROLLBAR_PRIVATE"))
	rollbar.SetEnvironment(envString)                  // defaults to "development"
	rollbar.SetCodeVersion("master")                   // optional Git hash/branch/tag (required for GitHub integration)
	rollbar.SetServerRoot("github.com/gamedb/website") // path of project (required for GitHub integration and non-project stacktrace collapsing)
}

func Log(interfaces ...interface{}) {

	interfaces = removeNils(interfaces...)

	if len(interfaces) == 0 {
		return
	}

	interfaces = addDefaultLogName(interfaces...)
	interfaces = addDefaultService(interfaces...)
	interfaces = addDefaultSeverity(interfaces...)

	var loggingServices []Service
	var googleLogs []*logging.Logger
	var entry logging.Entry

	// Create entry
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
			googleLogs = append(googleLogs, googleClient.Logger(string(val)+"-"+string(env)))
		case Severity:
			entry.Severity = val
		case time.Time:
			entry.Timestamp = val
		case Service:
			loggingServices = append(loggingServices, val)
		default:
			Log("Invalid value given to Err")
		}
	}

	if entry.Payload.(string) == "" {
		return
	}

	// Default log
	if len(googleLogs) == 0 {
		googleLogs = append(googleLogs, googleClient.Logger(string(LogNameGameDB)+"-"+string(env)))
	}

	// Add stack to payload
	//entry.Payload = entry.Payload.(string) + "\n\r" + string(debug.Stack())

	for _, v := range loggingServices {
		if v == ServiceGoogle {
			for _, vv := range googleLogs {
				vv.Log(entry)
			}
		}
		if v == ServiceLocal {
			logger.Println(entry.Payload.(string))
		}
		if v == ServiceRollbar {

			switch entry.Severity {
			case SeverityCritical:
				rollbar.Critical(entry.Payload)
			default:
			case SeverityError:
				rollbar.Error(entry.Payload)
			case SeverityWarning:
				rollbar.Warning(entry.Payload)
			case SeverityInfo:
				rollbar.Info(entry.Payload)
			case SeverityDebug:
				rollbar.Debug(entry.Payload)
			}
		}
	}
}

func addDefaultService(interfaces ...interface{}) []interface{} {

	for _, v := range interfaces {
		_, ok := v.(Service)
		if ok {
			return interfaces
		}
	}

	return append(interfaces, ServiceGoogle, ServiceLocal)
}

func addDefaultSeverity(interfaces ...interface{}) []interface{} {

	for _, v := range interfaces {
		_, ok := v.(Severity)
		if ok {
			return interfaces
		}
	}

	return append(interfaces, SeverityError)
}

func addDefaultLogName(interfaces ...interface{}) []interface{} {

	for _, v := range interfaces {
		_, ok := v.(LogName)
		if ok {
			return interfaces
		}
	}

	return append(interfaces, LogNameGameDB)
}

func removeNils(interfaces ...interface{}) (ret []interface{}) {

	for _, v := range interfaces {
		if v != nil {
			ret = append(ret, v)
		}
	}

	return ret
}
