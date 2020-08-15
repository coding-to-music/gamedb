package steam

import (
	"github.com/gamedb/gamedb/pkg/config"
	"go.uber.org/zap"
)

type steamLogger struct {
}

func (l steamLogger) Info(s string) {
	if config.IsLocal() {
		// zap.S().Info(s, log.LogNameSteam)
	}
}

func (l steamLogger) Err(e error) {
	zap.S().Error(e)
}
