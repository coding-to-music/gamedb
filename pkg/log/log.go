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

		cores = []zapcore.Core{
			getStandardCore(encoderConfig),
			newFileCore(encoderConfig, logName),
		}

	} else {

		options = append(options, zap.AddStacktrace(zap.WarnLevel))

		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = ""
		encoderConfig.LevelKey = ""

		cores = []zapcore.Core{
			getStandardCore(encoderConfig),
		}

		if config.C.GoogleProject != "" && config.C.GoogleAuthFile != "" {
			cores = append(cores, newGoogleCore(encoderConfig))
		}

		if config.C.RollbarSecret != "" && config.C.RollbarUser != "" {
			cores = append(cores, newRollbarCore(false))
		}
	}

	logger := zap.New(zapcore.NewTee(cores...), options...).Named(logName)

	zap.ReplaceGlobals(logger)
}

func Flush() {

	recover()

	if err := zap.L().Sync(); err != nil {

		// Ignore unactionable errors
		if osErr, ok := err.(*os.PathError); ok {
			switch osErr.Unwrap() {
			case syscall.EINVAL, syscall.ENOTSUP, syscall.ENOTTY:
				return
			}
		}

		zap.L().Error("Failed to sync logs", zap.Error(err))
	}
}

func getStandardCore(encoderConfig zapcore.EncoderConfig) zapcore.Core {

	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	output := zapcore.Lock(os.Stdout)
	level := zap.NewAtomicLevelAt(zapcore.DebugLevel)

	return zapcore.NewCore(encoder, output, level)
}
