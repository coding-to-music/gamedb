package steam

import (
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
)

type steamLogger struct {
}

func (l steamLogger) Info(s string) {
	if config.IsLocal() {
		// log.Info(s, log.LogNameSteam)
	}
}

func (l steamLogger) Err(e error) {
	log.Err(e)
}
