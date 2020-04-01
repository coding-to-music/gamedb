package twitch

import (
	"sync"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/nicklaw5/helix"
)

var client *helix.Client
var lock sync.Mutex

func GetTwitch() (*helix.Client, error) {

	lock.Lock()
	defer lock.Unlock()

	var err error

	if client == nil {
		client, err = helix.NewClient(&helix.Options{
			ClientID:     config.Config.TwitchClientID.Get(),
			ClientSecret: config.Config.TwitchClientSecret.Get(),
		})
	}

	return client, err
}
