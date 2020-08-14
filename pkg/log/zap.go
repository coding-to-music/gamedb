package log

import (
	"context"

	"cloud.google.com/go/logging"
	"github.com/gamedb/gamedb/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitZap(logName LogName) {

	var logger *zap.Logger

	if config.IsLocal() {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}

	googleClient, err := logging.NewClient(context.Background(), config.Config.GoogleProject.Get())
	if err != nil {
		zap.S().Error(err.Error())
	}

	logger = logger.WithOptions(zap.Hooks(func(e zapcore.Entry) error {
		if !config.IsLocal() {
			if googleClient != nil {

				var level logging.Severity
				var message = e.Message

				if e.Level > zapcore.InfoLevel {
					message += "\n" + e.Stack
				}

				switch e.Level {
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

				googleClient.Logger(string(logName)).Log(logging.Entry{
					Timestamp: e.Time,
					Severity:  level,
					Payload:   message,
					Labels: map[string]string{
						"env":    config.Config.Environment.Get(),
						"commit": config.Config.CommitHash.Get(),
						"key":    config.Config.SteamAPIKey.Get(),
						"ip":     config.Config.IP.Get(),
					},
				})
			}
		}
		return nil
	}))

	zap.ReplaceGlobals(logger)
}
