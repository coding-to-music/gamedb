package log

import (
	"context"
	"fmt"
	l "log"
	"net"
	"net/http"
	"os"
	"path"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/logging"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/logrusorgru/aurora"
	"github.com/rollbar/rollbar-go"
)

//noinspection GoUnusedConst
const (
	// Log names
	LogNameConsumers LogName = "consumers"
	LogNameCron      LogName = "crons"
	LogNameDatastore LogName = "datastore"
	LogNameMongo     LogName = "mongo"
	LogNameDebug     LogName = "debug"
	LogNameGameDB    LogName = "gamedb" // Default
	LogNameRequests  LogName = "requests"
	LogNameSQL       LogName = "sql"
	LogNameSteam     LogName = "steam-calls"

	// Severities
	SeverityDebug    Severity = "debug"
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error" // Default
	SeverityCritical Severity = "critical"
)

type LogName string
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

type entry struct {
	request   *http.Request
	texts     []string
	error     string
	logName   LogName
	severity  Severity
	timestamp time.Time
}

func (e entry) toText(includeStack bool) string {

	var ret []string

	// Severity
	ret = append(ret, strings.ToUpper(string(e.severity)))

	// Environment
	if !config.IsLocal() {
		ret = append(ret, strings.ToUpper(config.Config.Environment.Get()))
		ret = append(ret, strings.ToUpper(path.Base(os.Args[0])))
	}

	// Request
	if e.request != nil {
		ret = append(ret, e.request.Method+" "+e.request.URL.Path)
	}

	// Texts
	ret = append(ret, e.texts...)

	// Error
	if e.error != "" {
		ret = append(ret, e.error)
	}

	// Join
	str := strings.Join(ret, " - ")

	// Stack
	if includeStack {
		str += "\n" + string(debug.Stack())
	}

	return str
}

var (
	googleClient *logging.Client
	logger       = l.New(os.Stderr, "", l.Ltime)
)

func init() {
	var err error
	googleClient, err = logging.NewClient(context.Background(), config.Config.GoogleProject.Get())
	if err != nil {
		fmt.Println(err)
	}

	rollbar.SetToken(config.Config.RollbarSecret.Get())
	rollbar.SetEnvironment(config.Config.Environment.Get())
	rollbar.SetCodeVersion(config.Config.CommitHash.Get())
	rollbar.SetServerHost("gamedb.online")
	rollbar.SetServerRoot("github.com/gamedb/gamedb")
}

func log(interfaces ...interface{}) {

	var entry = entry{
		logName:   LogNameGameDB,
		severity:  SeverityError,
		timestamp: time.Now(),
	}
	var loggingServices []Service

	// Create entry
	for _, v := range interfaces {

		switch val := v.(type) {
		case nil:
			continue
		case []byte:
			entry.texts = append(entry.texts, string(val))
		case net.IP:
			entry.texts = append(entry.texts, string(val))
		case bool:
			entry.texts = append(entry.texts, strconv.FormatBool(val))
		case time.Duration:
			entry.texts = append(entry.texts, val.String())
		case int:
			entry.texts = append(entry.texts, strconv.Itoa(val))
		case int64:
			entry.texts = append(entry.texts, strconv.FormatInt(val, 10))
		case float32:
			entry.texts = append(entry.texts, strconv.FormatFloat(float64(val), 'f', -1, 32))
		case float64:
			entry.texts = append(entry.texts, strconv.FormatFloat(float64(val), 'f', -1, 64))
		case string:
			entry.texts = append(entry.texts, val)
		case *http.Request:
			entry.request = val
		case error:
			entry.error = val.Error()
		case LogName:
			entry.logName = val
		case Severity:
			entry.severity = val
		case time.Time:
			entry.timestamp = val
		case Service:
			loggingServices = append(loggingServices, val)
		case Option:
		default:
			Warning("Invalid value given to log: " + reflect.TypeOf(val).String())
		}
	}

	if len(entry.texts) > 0 || entry.error != "" {

		// Local
		switch entry.severity {
		case SeverityCritical:
			logger.Println(aurora.Red(aurora.Bold(entry.toText(true))))
		case SeverityError:
			logger.Println(aurora.Red(entry.toText(true)))
		case SeverityWarning:
			logger.Println(aurora.Yellow(entry.toText(true)))
		case SeverityInfo:
			logger.Println(entry.toText(false))
		case SeverityDebug:
			logger.Println(aurora.Green(entry.toText(false)))
		default:
			logger.Println(entry.toText(false))
		}

		if !config.IsLocal() {

			// Google
			googleClient.Logger(config.Config.Environment.Get() + "-" + string(entry.logName)).Log(logging.Entry{
				Severity:  entry.severity.toGoole(),
				Timestamp: entry.timestamp,
				Payload:   entry.toText(true),
				Labels: map[string]string{
					"env": config.Config.Environment.Get(),
					"key": config.GetSteamKeyTag(),
				},
			})

			// Rollbar
			if entry.severity == SeverityWarning || entry.severity == SeverityError || entry.severity == SeverityCritical {
				rollbar.Log(rollbar.ERR, entry.toText(false))
			}
		}
	}
}

func Critical(interfaces ...interface{}) {
	log(append(interfaces, SeverityCritical)...)
}

func Err(interfaces ...interface{}) {
	log(append(interfaces, SeverityError)...)
}

func Warning(interfaces ...interface{}) {
	log(append(interfaces, SeverityWarning)...)
}

func Info(interfaces ...interface{}) {
	log(append(interfaces, SeverityInfo)...)
}

func Debug(interfaces ...interface{}) {
	log(append(interfaces, SeverityDebug)...)
}
