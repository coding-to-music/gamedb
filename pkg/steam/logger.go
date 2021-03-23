package steam

import (
	"github.com/gamedb/gamedb/pkg/log"
	"go.uber.org/zap"
)

type steamLogger struct {
}

func (l steamLogger) Info(s string) {
	// if config.IsLocal() {
	// 	zap.S().Named(log.LogNameSteamErrors).Info(s)
	// }
}

func (l steamLogger) Err(e error) {
	if e != nil {
		zap.S().Named(log.LogNameSteamErrors).Error(e)
	}
}
