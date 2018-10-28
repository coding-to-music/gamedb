package helpers

import (
	"github.com/Jleagle/steam-go/steam"
	"github.com/spf13/viper"
)

var steamClient *steam.Steam

func GetSteam() (*steam.Steam) {

	if steamClient == nil {

		steamClient = &steam.Steam{
			Key:        viper.GetString("API_KEY"),
			LogChannel: GetSteamLogsChan(),
		}

		steamClient.SetRateLimit(1)
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
