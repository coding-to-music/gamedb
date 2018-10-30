package helpers

import (
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/spf13/viper"
)

var steamClient *steam.Steam

func GetSteam() (*steam.Steam) {

	if steamClient == nil {

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

	return steamClient
}

var steamLogs chan string

func GetSteamLogsChan() chan string {

	if steamLogs == nil {
		steamLogs = make(chan string)
	}

	return steamLogs
}
