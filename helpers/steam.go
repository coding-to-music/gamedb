package helpers

import (
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/config"
)

var steamClient *steam.Steam

func GetSteam() *steam.Steam {

	if steamClient == nil {

		steamClient = new(steam.Steam)
		steamClient.SetKey(config.Config.SteamAPIKey)
		steamClient.SetUserAgent("http://gamedb.online")
		steamClient.SetLogChannel(GetSteamLogsChan())
		steamClient.SetAPIRateLimit(time.Millisecond*1000, 10)
		steamClient.SetStoreRateLimit(time.Millisecond*1600, 10)

	}

	return steamClient
}

var steamLogs chan steam.Log

func GetSteamLogsChan() chan steam.Log {

	if steamLogs == nil {
		steamLogs = make(chan steam.Log)
	}

	return steamLogs
}
