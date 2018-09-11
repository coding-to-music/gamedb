package steami

import (
	"github.com/Jleagle/steam-go/steam"
	"github.com/spf13/viper"
)

var steamClient *steam.Steam
var steamLogs chan string

func Steam() (*steam.Steam) {

	if steamClient == nil {

		s := steam.Steam{
			Key:        viper.GetString("API_KEY"),
			LogChannel: steamLogs,
			Throttle:   false, // todo, this doesnt work!
			Format:     "json",
		}

		steamClient = &s
	}

	return steamClient
}
