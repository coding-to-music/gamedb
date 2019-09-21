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
	"github.com/getsentry/sentry-go"
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
	LogNamePICS      LogName = "pics-checks"

	// Severities
	SeverityDebug    Severity = 1
	SeverityInfo     Severity = 2
	SeverityWarning  Severity = 3
	SeverityError    Severity = 4 // Default
	SeverityCritical Severity = 5
)

type LogName string
type Service string
type Option int
type Severity int

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

func (s Severity) string() string {

	switch s {
	case SeverityDebug:
		return "Debug"
	case SeverityInfo:
		return "Info"
	case SeverityWarning:
		return "Warning"
	case SeverityError:
		return "Error"
	case SeverityCritical:
		return "Critical"
	default:
		return "Error"
	}
}

type entry struct {
	request   *http.Request
	texts     []string
	error     error
	logName   LogName
	severity  Severity
	timestamp time.Time
}

func (e entry) toText(severity Severity) string {

	var ret []string

	// Severity
	ret = append(ret, e.severity.string())

	// Environment
	if !config.IsLocal() {
		ret = append(ret, config.Config.Environment.Get())
		ret = append(ret, path.Base(os.Args[0]))
	}

	// Request
	if e.request != nil {
		ret = append(ret, e.request.Method+" "+e.request.URL.Path)
	}

	// Texts
	ret = append(ret, e.texts...)

	// Error
	if e.error != nil {
		ret = append(ret, e.error.Error())
	}

	// Join
	str := strings.Join(ret, " - ")

	// Stack
	if severity > 3 {
		str += "\n" + string(debug.Stack())
	}

	return str
}

var (
	googleClient *logging.Client
	logger       = l.New(os.Stderr, "", l.Ltime)
)

func Initialise() {

	var err error

	// Google
	googleClient, err = logging.NewClient(context.Background(), config.Config.GoogleProject.Get())
	if err != nil {
		fmt.Println(err)
	}

	// Rollbar
	rollbar.SetToken(config.Config.RollbarSecret.Get())
	rollbar.SetEnvironment(config.Config.Environment.Get())
	rollbar.SetServerHost("gamedb.online")
	rollbar.SetServerRoot("github.com/gamedb/gamedb")

	// Sentry
	err = sentry.Init(sentry.ClientOptions{
		Dsn:              config.Config.SentryDSN.Get(),
		AttachStacktrace: true,
		Environment:      config.Config.Environment.Get(),
		Release:          config.Config.CommitHash.Get(),
	})
	if err != nil {
		fmt.Println(err)
	}
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
		case []string:
			entry.texts = append(entry.texts, strings.Join(val, ","))
		case bool:
			entry.texts = append(entry.texts, strconv.FormatBool(val))
		case time.Duration:
			entry.texts = append(entry.texts, val.String())
		case int:
			entry.texts = append(entry.texts, strconv.Itoa(val))
		case uint32:
			entry.texts = append(entry.texts, strconv.FormatUint(uint64(val), 10))
		case uint64:
			entry.texts = append(entry.texts, strconv.FormatUint(val, 10))
		case int64:
			entry.texts = append(entry.texts, strconv.FormatInt(val, 10))
		case float32:
			entry.texts = append(entry.texts, strconv.FormatFloat(float64(val), 'f', -1, 32))
		case float64:
			entry.texts = append(entry.texts, strconv.FormatFloat(val, 'f', -1, 64))
		case string:
			entry.texts = append(entry.texts, val)
		case *http.Request:
			entry.request = val
		case error:
			entry.error = val
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

	if len(entry.texts) > 0 || entry.error != nil {

		// Local
		switch entry.severity {
		case SeverityCritical:
			logger.Println(aurora.Red(aurora.Bold(entry.toText(entry.severity))))
		case SeverityError:
			logger.Println(aurora.Red(entry.toText(entry.severity)))
		case SeverityWarning:
			logger.Println(aurora.Yellow(entry.toText(entry.severity)))
		case SeverityInfo:
			logger.Println(entry.toText(entry.severity))
		case SeverityDebug:
			logger.Println(aurora.Green(entry.toText(entry.severity)))
		default:
			logger.Println(entry.toText(entry.severity))
		}

		if !config.IsLocal() {

			// Google
			googleClient.Logger(config.Config.Environment.Get() + "-" + string(entry.logName)).Log(logging.Entry{
				Severity:  entry.severity.toGoole(),
				Timestamp: entry.timestamp,
				Payload:   entry.toText(entry.severity),
				Labels: map[string]string{
					"env":  config.Config.Environment.Get(),
					"key":  config.GetSteamKeyTag(),
					"hash": config.Config.CommitHash.Get(),
				},
			})

			if entry.severity == SeverityWarning || entry.severity == SeverityError || entry.severity == SeverityCritical {

				// Rollbar
				rollbar.Log(rollbar.ERR, entry.toText(SeverityInfo))

				// Sentry
				sentry.CaptureException(entry.error)
				sentry.Flush(time.Second * 5)
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
