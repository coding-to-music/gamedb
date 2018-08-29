package steami

import (
	"os"

	"github.com/Jleagle/steam-go/steam"
)

var steamClient *steam.Steam
var steamLogs chan string

func Steam() (*steam.Steam) {

	if steamClient == nil {

		s := steam.Steam{
			Key:        os.Getenv("STEAM_API_KEY"),
			LogChannel: steamLogs,
			Throttle:   false, // todo, this doesnt work!
			Format:     "json",
		}

		steamClient = &s
	}

	return steamClient
}
