package steam

import (
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/config"
)

type steamLogger struct {
}

func (l steamLogger) Write(i steamapi.Log) {
	if config.IsLocal() {
		// log.Info(i.String(), log.LogNameSteam)
	}
}
