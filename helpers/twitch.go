package helpers

import (
	"sync"

	"github.com/gamedb/website/config"
	"github.com/nicklaw5/helix"
)

var twitchClient *helix.Client
var twitchLock sync.Mutex

func GetTwitch() (client *helix.Client, err error) {

	twitchLock.Lock()
	defer twitchLock.Unlock()

	if twitchClient == nil {
		twitchClient, err = helix.NewClient(&helix.Options{
			ClientID: config.Config.TwitchClientID.Get(),
		})
	}

	return twitchClient, err
}
