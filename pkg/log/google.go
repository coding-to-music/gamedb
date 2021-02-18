package log

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/logging"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/ldflags"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/api/option"
)

func newGoogleCore(encoderConfig zapcore.EncoderConfig) zapcore.Core {

	ctx := context.Background()

	googleClient, err := logging.NewClient(ctx, config.C.GoogleProject, option.WithCredentialsFile(config.C.GoogleAuthFile))
	if err != nil {
		fmt.Println(err)
	}

	return googleCore{
		client:  googleClient,
		context: ctx,
		loggers: map[string]*logging.Logger{},
		async:   true,

		levelEnabler: zap.NewAtomicLevelAt(zapcore.DebugLevel),
		encoder:      zapcore.NewConsoleEncoder(googleEncoderConfig()),
		output:       zapcore.AddSync(io.Discard),
	}
}

type googleCore struct {
	client  *logging.Client
	context context.Context
	loggers map[string]*logging.Logger
	async   bool

	levelEnabler zapcore.LevelEnabler
	encoder      zapcore.Encoder
	output       zapcore.WriteSyncer
}

func (g *googleCore) clone() *googleCore {

	return &googleCore{
		client:  g.client,
		context: g.context,
		loggers: g.loggers,
		async:   g.async,

		levelEnabler: g.levelEnabler,
		encoder:      g.encoder.Clone(),
		output:       g.output,
	}
}

func (g *googleCore) getLogger(name string) *logging.Logger {

	// Return cached logger
	// Disabled for now to get up to date common values
	// if val, ok := g.loggers[name]; ok {
	// 	return val
	// }

	// Create logger
	common := map[string]string{
		"env":     config.C.Environment,
		"key":     config.C.SteamAPIKey,
		"ip":      config.C.IP,
		"commits": ldflags.CommitCount,
		"hash":    ldflags.CommitHash,
	}
	g.loggers[name] = g.client.Logger(name, logging.CommonLabels(common))

	return g.loggers[name]
}

func (g googleCore) Enabled(level zapcore.Level) bool {
	return level.Enabled(level)
}

func (g googleCore) With(fields []zapcore.Field) zapcore.Core {

	clone := g.clone()
	for k := range fields {
		fields[k].AddTo(clone.encoder)
	}
	return clone
}

func (g googleCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {

	if g.Enabled(entry.Level) {
		return checkedEntry.AddCore(entry, g)
	}
	return checkedEntry
}

func (g googleCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {

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

func (g googleCore) Sync() error {

	for _, logger := range g.loggers {

		err := logger.Flush()
		if err != nil {
			return err
		}
	}

	return g.output.Sync()
}

// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry
func googleEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:       "timestamp",
		LevelKey:      "severity",
		NameKey:       "logName",
		CallerKey:     "caller",
		MessageKey:    "textPayload",
		StacktraceKey: "trace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel: func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			switch l {
			case zapcore.DebugLevel:
				enc.AppendString("DEBUG")
			case zapcore.InfoLevel:
				enc.AppendString("INFO")
			case zapcore.WarnLevel:
				enc.AppendString("WARNING")
			case zapcore.ErrorLevel:
				enc.AppendString("ERROR")
			case zapcore.DPanicLevel:
				enc.AppendString("CRITICAL")
			case zapcore.PanicLevel:
				enc.AppendString("ALERT")
			case zapcore.FatalLevel:
				enc.AppendString("EMERGENCY")
			}
		},
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}
