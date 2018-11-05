package helpers

import (
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/spf13/viper"
)

var steamClient *steam.Steam

// Called from main
func Init() {

	steamClient = &steam.Steam{
		Key:        viper.GetString("API_KEY"),
		LogChannel: GetSteamLogsChan(),
		UserAgent:  "http://gamedb.online",
	}

	steamClient.SetRateLimit(
		time.Hour*24/100000,
		time.Millisecond*1600,
	)
}

func GetSteam() (*steam.Steam) {
	return steamClient
}

var steamLogs chan steam.Log

func GetSteamLogsChan() chan steam.Log {

	if steamLogs == nil {
		steamLogs = make(chan steam.Log)
	}

	return steamLogs
}
