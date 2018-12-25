package helpers

import (
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/config"
)

var steamClient = &steam.Steam{
	Key:        config.Config.SteamAPIKey,
	LogChannel: GetSteamLogsChan(),
	UserAgent:  "http://gamedb.online",
	APIRate:    time.Millisecond * 1000,
	StoreRate:  time.Millisecond * 1600,
}

func GetSteam() *steam.Steam {
	return steamClient
}

var steamLogs chan steam.Log

func GetSteamLogsChan() chan steam.Log {

	if steamLogs == nil {
		steamLogs = make(chan steam.Log)
	}

	return steamLogs
}
