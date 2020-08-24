package log

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"cloud.google.com/go/logging"
	"github.com/gamedb/gamedb/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//goland:noinspection GoUnusedConst
const (
	// Binaries
	LogNameAPI       = "api"
	LogNameBackend   = "backend"
	LogNameChatbot   = "chatbot"
	LogNameConsumers = "consumers"
	LogNameCrons     = "crons"
	LogNameSteam     = "steam"
	LogNameFrontend  = "frontend"
	LogNameTest      = "test"
	LogNameScaler    = "scaler"

	// Others
	LogNameMongo          = "mongo"
	LogNameRabbit         = "rabbit"
	LogNameRequests       = "requests"
	LogNameSQL            = "sql"
	LogNameTriggerUpdate  = "trigger"
	LogNameSteamErrors    = "steam-lib"
	LogNameWebhooksGitHub = "github"
	LogNameInflux         = "influx"
)

func InitZap(logName string) {

	var logger *zap.Logger

	if config.IsLocal() {
		logger = zap.New(
			getStandardCore(),
			zap.AddStacktrace(zap.WarnLevel),
			zap.AddCaller(),
			zap.Development(),
		)
	} else {
		logger = zap.New(
			zapcore.NewTee(getStandardCore(), getGoogleCore()),
			zap.AddStacktrace(zap.WarnLevel),
			zap.AddCaller(),
		)
	}

	logger = logger.Named(logName)

	zap.ReplaceGlobals(logger)
}

func getStandardCore() zapcore.Core {

	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	output := zapcore.Lock(os.Stdout)
	level := zap.NewAtomicLevelAt(zapcore.DebugLevel)

	return zapcore.NewCore(encoder, output, level)
}

func getGoogleCore() zapcore.Core {

	ctx := context.Background()

	googleClient, err := logging.NewClient(ctx, config.C.GoogleProject)
	if err != nil {
		fmt.Println(err)
	}

	return GoogleCore{
		client:  googleClient,
		context: ctx,
		loggers: map[string]*logging.Logger{},
		async:   true,

		levelEnabler: zap.NewAtomicLevelAt(zapcore.DebugLevel),
		encoder:      zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		output:       zapcore.AddSync(ioutil.Discard),
	}
}

type GoogleCore struct {
	client  *logging.Client
	context context.Context
	loggers map[string]*logging.Logger
	async   bool

	levelEnabler zapcore.LevelEnabler
	encoder      zapcore.Encoder
	output       zapcore.WriteSyncer
}

func (g *GoogleCore) clone() *GoogleCore {

	return &GoogleCore{
		client:  g.client,
		context: g.context,
		loggers: g.loggers,
		async:   g.async,

		levelEnabler: g.levelEnabler,
		encoder:      g.encoder.Clone(),
		output:       g.output,
	}
}

func (g *GoogleCore) getLogger(name string) *logging.Logger {

	// Return cached logger
	if val, ok := g.loggers[name]; ok {
		return val
	}

	// Create logger
	common := map[string]string{
		"env":     config.C.Environment,
		"commits": config.C.Commits,
		"hash":    config.C.CommitHash,
		"key":     config.C.SteamAPIKey,
		"ip":      config.C.IP,
	}
	g.loggers[name] = g.client.Logger(name, logging.CommonLabels(common))

	return g.loggers[name]
}

func (g GoogleCore) Enabled(level zapcore.Level) bool {
	return level.Enabled(level)
}

func (g GoogleCore) With(fields []zapcore.Field) zapcore.Core {

	clone := g.clone()
	for k := range fields {
		fields[k].AddTo(clone.encoder)
	}
	return clone
}

func (g GoogleCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {

	if g.Enabled(entry.Level) {
		return checkedEntry.AddCore(entry, g)
	}
	return checkedEntry
}

func (g GoogleCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {

	buf, err := g.encoder.EncodeEntry(entry, fields)
	if err != nil {
		return err
	}

	var level logging.Severity

	switch entry.Level {
	case zapcore.DebugLevel:
		level = logging.Debug
	case zapcore.InfoLevel:
		level = logging.Info
	case zapcore.WarnLevel:
		level = logging.Warning
	case zapcore.ErrorLevel:
		level = logging.Error
	case zapcore.DPanicLevel:
		level = logging.Critical
	case zapcore.PanicLevel:
		level = logging.Critical
	default:
		level = logging.Debug
	}

	googleEntry := logging.Entry{
		Timestamp: entry.Time,
		Severity:  level,
		Payload:   buf.String(),
	}

	logger := g.getLogger(entry.LoggerName)

	if g.async {
		logger.Log(googleEntry)
	} else {
		err = logger.LogSync(g.context, googleEntry)
	}

	return err
}

func (g GoogleCore) Sync() error {

	for _, logger := range g.loggers {

		err := logger.Flush()
		if err != nil {
			return err
		}
	}

	return g.output.Sync()
}
