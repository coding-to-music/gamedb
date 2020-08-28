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
	LogNameTest      = "test"
	LogNameScaler    = "scaler"

	// Others
	LogNameMongo          = "mongo"
	LogNameTwitter        = "twitter" // Zapier
	LogNameRabbit         = "rabbit"
	LogNameRequests       = "requests"
	LogNameSQL            = "sql"
	LogNameTriggerUpdate  = "trigger"
	LogNameSteamErrors    = "steam-lib"
	LogNameWebhooksGitHub = "github"
	LogNameInflux         = "influx"
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
