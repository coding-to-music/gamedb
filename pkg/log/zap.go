package log

import (
	"os"
	"syscall"

	"github.com/gamedb/gamedb/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitZap(logName string) {

	var cores []zapcore.Core

	var options = []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	}

	if config.IsLocal() {

		options = append(options, zap.Development())

		encoderConfig := zap.NewDevelopmentEncoderConfig()

		cores = append(cores,
			getStandardCore(encoderConfig),
			newFileCore(encoderConfig, logName),
		)

	} else {

		options = append(options, zap.AddStacktrace(zap.WarnLevel))

		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = ""
		encoderConfig.LevelKey = ""

		cores = append(cores,
			getStandardCore(encoderConfig),
		)

		if config.C.GoogleProject != "" && config.C.GoogleAuthFile != "" {
			cores = append(cores, newGoogleCore(encoderConfig))
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

	err := recover()
	if err != nil {
		zap.S().Error(err)
	}

	err = zap.L().Sync()

	// Ignore unactionable errors
	if osErr, ok := err.(*os.PathError); ok {
		wrappedErr := osErr.Unwrap()
		switch wrappedErr {
		case syscall.EINVAL, syscall.ENOTSUP, syscall.ENOTTY:
			err = nil
		}
	}

	if err != nil {
		zap.S().Error(err)
	}
}

func getStandardCore(encoderConfig zapcore.EncoderConfig) zapcore.Core {

	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	output := zapcore.Lock(os.Stdout)
	level := zap.NewAtomicLevelAt(zapcore.DebugLevel)

	return zapcore.NewCore(encoder, output, level)
}
