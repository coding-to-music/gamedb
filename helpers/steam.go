package helpers

import (
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/spf13/viper"
)

var steamClient *steam.Steam

func GetSteam() (*steam.Steam) {

	time.Sleep(time.Second * 2) // Temporary

	if steamClient == nil {

		s := steam.Steam{
			Key:        viper.GetString("API_KEY"),
			LogChannel: GetSteamLogsChan(),
			Throttle:   false, // todo, this doesnt work!
			Format:     "json",
		}

		steamClient = &s
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
