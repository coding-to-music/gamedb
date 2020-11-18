package log

import (
	"go.uber.org/zap"
)

const (
	// Binaries
	LogNameAPI       = "api"
	LogNameBackend   = "backend"
	LogNameChatbot   = "chatbot"
	LogNameConsumers = "consumers"
	LogNameCrons     = "crons"
	LogNameSteam     = "steam"
	LogNameFrontend  = "frontend"
	LogNameDown      = "down"
	LogNameTest      = "test"
	LogNameScaler    = "scaler"
	LogNameDevenv    = "devenv"

	// Webhooks
	LogNameTwitter  = "twitter" // Zapier
	LogNameSendGrid = "sendgrid"
	LognameGitHub   = "github"

	// Others
	LogNameMongo         = "mongo"
	LogNameRabbit        = "rabbit"
	LogNameRequests      = "requests"
	LogNameSQL           = "sql"
	LogNameTriggerUpdate = "trigger"
	LogNameSteamErrors   = "steam-lib"
	LogNameInflux        = "influx"
)

func Debug(msg string, fields ...zap.Field) {
	zap.L().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	zap.L().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	zap.L().Warn(msg, fields...)
}

func Err(msg string, fields ...zap.Field) {
	zap.L().Error(msg, fields...)
}

// The logger then calls os.Exit(1)
func Fatal(msg string, fields ...zap.Field) {
	zap.L().Fatal(msg, fields...)
}

func DebugS(args ...interface{}) {
	zap.S().Debug(args...)
}

func InfoS(args ...interface{}) {
	zap.S().Info(args...)
}

func WarnS(args ...interface{}) {
	zap.S().Warn(args...)
}

func ErrS(args ...interface{}) {
	zap.S().Error(args...)
}

func FatalS(args ...interface{}) {
	zap.S().Fatal(args...)
}
