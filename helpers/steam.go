package helpers

import (
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/log"
)

var steamClient *steam.Steam

func GetSteam() *steam.Steam {

	if steamClient == nil {

		steamClient = new(steam.Steam)
		steamClient.SetKey(config.Config.SteamAPIKey)
		steamClient.SetUserAgent("http://gamedb.online")
		steamClient.SetLogger(steamLogger{})
		steamClient.SetAPIRateLimit(time.Millisecond*1000, 10)
		steamClient.SetStoreRateLimit(time.Millisecond*1600, 10)

	}

	return steamClient
}

type steamLogger struct {
}

func (l steamLogger) Write(i steam.Log) {
	if config.Config.IsLocal() {
		log.Info(i.String(), log.LogNameSteam)
	}
}
