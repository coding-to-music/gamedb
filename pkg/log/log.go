package log

import (
	"context"
	"fmt"
	logg "log"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/logging"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/logrusorgru/aurora"
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
	text      string
	error     string
	logName   LogName
	severity  Severity
	timestamp time.Time
}

func (e entry) toText(includeStack bool) string {

	var ret []string

	ret = append(ret, strings.ToUpper(string(e.severity)))

	if e.request != nil {
		ret = append(ret, e.request.Method+" "+e.request.URL.Path)
	}

	if e.text != "" {
		ret = append(ret, e.text)
	}

	if e.error != "" {
		ret = append(ret, e.error)
	}

	str := strings.Join(ret, " - ")

	if includeStack {
		str += "\n" + string(debug.Stack())
	}

	return str
}

var (
	googleClient *logging.Client
	logger       = logg.New(os.Stderr, "", logg.Ltime)
)

func init() {
	var err error
	googleClient, err = logging.NewClient(context.Background(), config.Config.GoogleProject.Get())
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
		case time.Duration:
			entry.text = val.String()
		case int:
			entry.text = strconv.Itoa(val)
		case int64:
			entry.text = strconv.FormatInt(val, 10)
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
		default:
			Warning("Invalid value given to log: " + reflect.TypeOf(val).String())
		}
	}

	if entry.text == "" && entry.error == "" {
		return
	}

	switch entry.severity {
	case SeverityCritical:
		logger.Println(aurora.Red(aurora.Bold(entry.toText(true))))
	case SeverityError:
		logger.Println(aurora.Red(entry.toText(true)))
	case SeverityWarning:
		logger.Println(aurora.Brown(entry.toText(true)))
	case SeverityInfo:
		logger.Println(entry.toText(false))
	case SeverityDebug:
		logger.Println(aurora.Green(entry.toText(false)))
	default:
		logger.Println(entry.toText(false))
	}

	if config.IsProd() {

		googleClient.Logger(config.Config.Environment.Get() + "-" + string(entry.logName)).Log(logging.Entry{
			Severity:  entry.severity.toGoole(),
			Timestamp: entry.timestamp,
			Payload:   entry.toText(true),
			Labels: map[string]string{
				"env": config.Config.Environment.Get(),
				"key": config.GetSteamKeyTag(),
			},
		})
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
