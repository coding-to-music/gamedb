package log

import (
	"context"
	"fmt"
	l "log"
	"net"
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

type LogName string

const (
	// Binaries
	LogNameChatbot   LogName = "binary-chatbot"
	LogNameConsumers LogName = "binary-consumers"
	LogNameCrons     LogName = "binary-crons"
	LogNameSteam     LogName = "binary-steam"
	LogNameWebserver LogName = "binary-webserver"
	LogNameTest      LogName = "binary-test"
	LogNameScaler    LogName = "binary-scaler"

	// Others
	LogNameMongo         LogName = "mongo"
	LogNameRabbit        LogName = "rabbit"
	LogNameRequests      LogName = "requests"
	LogNameSQL           LogName = "sql"
	LogNameTriggerUpdate LogName = "trigger-update"
	LogNameSteamErrors   LogName = "steam-errors"
	// LogNameInflux        LogName = "influx"
)

type entry struct {
	request  *http.Request
	texts    []string
	error    error
	severity logging.Severity
}

func (e entry) string(severity logging.Severity) string {

	var ret []string

	// Request
	if e.request != nil {
		ret = append(ret, e.request.Method+" "+e.request.URL.Path)
	}

	// Error
	if e.error != nil {
		ret = append(ret, e.error.Error())
	}

	// Texts
	ret = append(ret, e.texts...)

	// Join
	str := strings.Join(ret, " - ")

	// Stack
	if (!config.IsLocal() && severity > logging.Info) || severity > logging.Warning {
		str += "\n" + string(debug.Stack())
	}

	return str
}

var (
	googleClient *logging.Client
	googleLog    LogName
	logger       *l.Logger
)

func Initialise(log LogName) {

	googleLog = log

	if config.IsLocal() {
		logger = l.New(os.Stderr, "", 0)
	} else {
		logger = l.New(os.Stderr, "", l.LstdFlags)
	}

	// Google
	var err error
	googleClient, err = logging.NewClient(context.Background(), config.Config.GoogleProject.Get())
	if err != nil {
		fmt.Println(err)
	}
}

func log(interfaces ...interface{}) {

	var entry = entry{severity: logging.Error}

	for _, v := range interfaces {

		switch val := v.(type) {
		case nil:
			continue
		case error:
			entry.error = val
		case float32:
			entry.texts = append(entry.texts, strconv.FormatFloat(float64(val), 'f', -1, 32))
		case float64:
			entry.texts = append(entry.texts, strconv.FormatFloat(val, 'f', -1, 64))
		case []byte:
			entry.texts = append(entry.texts, string(val))
		case time.Time:
			entry.texts = append(entry.texts, val.String())
		case time.Duration:
			entry.texts = append(entry.texts, val.String())
		case net.IP:
			entry.texts = append(entry.texts, val.String())
		case *http.Request:
			entry.request = val
		case logging.Severity:
			entry.severity = val
		default:
			entry.texts = append(entry.texts, fmt.Sprint(val))
		}
	}

	if len(entry.texts) > 0 || entry.error != nil {

		var text = entry.string(entry.severity)

		// Local
		switch entry.severity {
		case logging.Debug:
			logger.Println(aurora.Green(text))
		case logging.Info:
			logger.Println(text)
		case logging.Warning:
			logger.Println(aurora.Yellow(text))
		case logging.Error:
			logger.Println(aurora.Red(text))
		case logging.Critical:
			logger.Println(aurora.Red(aurora.Bold(text)))
		default:
			logger.Println(text)
		}

		// Google
		if !config.IsLocal() {
			googleClient.Logger(string(googleLog)).Log(logging.Entry{
				Severity: entry.severity,
				Payload:  text,
				Labels: map[string]string{
					"env":    config.Config.Environment.Get(),
					"commit": config.Config.CommitHash.Get(),
					"key":    config.Config.SteamAPIKey.Get(),
					"ip":     config.Config.IP.Get(),
				},
			})
		}
	}
}

func Debug(interfaces ...interface{}) {
	log(append(interfaces, logging.Debug)...)
}

func Info(interfaces ...interface{}) {
	log(append(interfaces, logging.Info)...)
}

func Warning(interfaces ...interface{}) {
	log(append(interfaces, logging.Warning)...)
}

func Err(interfaces ...interface{}) {
	log(append(interfaces, logging.Error)...)
}

func Critical(interfaces ...interface{}) {
	log(append(interfaces, logging.Critical)...)
}
