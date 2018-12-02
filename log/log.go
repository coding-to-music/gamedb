package log

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"cloud.google.com/go/logging"
	"github.com/rollbar/rollbar-go"
	"github.com/spf13/viper"
)

type LogName string
type Environment string
type Service string
type Option int
type Severity string

func (s Severity) toGoole() (severity logging.Severity) {

	switch s {
	case SeverityDebug:
		return logging.Debug
	case SeverityInfo:
		return logging.Info
	case SeverityWarning:
		return logging.Warning
	case SeverityError:
		return logging.Error
	case SeverityCritical:
		return logging.Critical
	default:
		return logging.Error
	}
}

func (s Severity) toRollbar() (severity string) {

	switch s {
	case SeverityDebug:
		return rollbar.DEBUG
	case SeverityInfo:
		return rollbar.INFO
	case SeverityWarning:
		return rollbar.WARN
	case SeverityError:
		return rollbar.ERR
	case SeverityCritical:
		return rollbar.CRIT
	default:
		return rollbar.ERR
	}
}

type entry struct {
	request   *http.Request
	text      string
	error     string
	logName   LogName
	severity  Severity
	timestamp time.Time
	stack     string
}

func (e entry) toText() string {

	var ret []string

	ret = append(ret, strings.ToUpper(string(e.severity)))

	if e.request != nil {
		ret = append(ret, e.request.Method, e.request.URL.Path)
	}

	if e.text != "" {
		ret = append(ret, e.text)
	}

	if e.error != "" {
		ret = append(ret, e.error)
	}

	if e.stack != "" {
		ret = append(ret, e.stack)
	}

	return strings.Join(ret, " - ")
}

//noinspection GoUnusedConst
const (
	// Environments
	EnvProd  Environment = "production"
	EnvLocal Environment = "local"

	// Log names
	LogNameConsumers LogName = "gamedb.consumers"
	LogNameSteam     LogName = "gamedb.steam"
	LogNameGameDB    LogName = "gamedb" // Default

	// Severities
	SeverityDebug    Severity = "debug"
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error" // Default
	SeverityCritical Severity = "critical"

	// Services
	ServiceGoogle  Service = "google"  // Default
	ServiceRollbar Service = "rollbar" //
	ServiceLocal   Service = "local"   // Default

	// Options
	OptionStack Option = iota
)

var (
	env Environment

	// Google
	googleCtx    = context.Background()
	googleClient *logging.Client

	// Local
	logger = log.New(os.Stderr, "", log.Ltime)
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

	var entry entry
	var loggingServices []Service

	// Create entry
	for _, v := range interfaces {

		switch val := v.(type) {
		case string:
			entry.text = val
		case *http.Request:
			entry.request = val
		case error:
			if val != nil {
				entry.error = val.Error()
			}
		case LogName:
			entry.logName = val
		case Severity:
			entry.severity = val
		case time.Time:
			entry.timestamp = val
		case Service:
			loggingServices = append(loggingServices, val)
		case Option:
			if val == OptionStack {
				entry.stack = string(debug.Stack())
			}
		default:
			Log("Invalid value given to Err")
		}
	}

	if entry.text == "" && entry.error == "" {
		return
	}

	// Send entry
	for _, v := range loggingServices {

		// Google
		if v == ServiceGoogle {
			googleClient.Logger(string(env) + "-" + string(entry.logName)).Log(logging.Entry{
				Severity:  entry.severity.toGoole(),
				Timestamp: entry.timestamp,
				Payload:   entry.toText(),
			})
		}

		// Local
		if v == ServiceLocal {
			logger.Println(entry.toText())
		}

		// Rollbar
		if v == ServiceRollbar {
			rollbar.Log(entry.severity.toRollbar(), interfaces...)
		}
	}
}

func Info(interfaces ...interface{}) {
	Log(append(interfaces, SeverityInfo)...)
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
