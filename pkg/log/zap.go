package log

import (
	"os"

	"github.com/gamedb/gamedb/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitZap(logName string) {

	var cores = []zapcore.Core{
		getStandardCore(),
	}

	var options = []zap.Option{
		zap.AddStacktrace(zap.WarnLevel),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	}

	if config.IsLocal() {

		options = append(options, zap.Development())

		cores = append(cores, newFileCore(logName))

	} else {

		if config.C.GoogleProject != "" && config.C.GoogleAuthFile != "" {
			cores = append(cores, newGoogleCore())
		}
		if config.C.RollbarSecret != "" && config.C.RollbarUser != "" {
			// Add rollbar core
		}
		if config.C.SentryDSN != "" {
			// Add sentry core
		}
	}

	logger := zap.New(zapcore.NewTee(cores...), options...).Named(logName)

	zap.ReplaceGlobals(logger)
}

func Flush() {

	Info("Flushing logs")

	err := recover()
	if err != nil {
		zap.S().Error(err)
	}

	err = zap.L().Sync()
	if err != nil {
		zap.S().Error(err)
	}
}

func getStandardCore() zapcore.Core {

	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	output := zapcore.Lock(os.Stdout)
	level := zap.NewAtomicLevelAt(zapcore.DebugLevel)

	return zapcore.NewCore(encoder, output, level)
}
