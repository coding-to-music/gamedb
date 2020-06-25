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

		ops := &helix.Options{
			ClientID:     config.Config.TwitchClientID.Get(),
			ClientSecret: config.Config.TwitchClientSecret.Get(),
		}

		client, err = helix.NewClient(ops)
		if err != nil {
			return nil, err
		}

		token, err := client.GetAppAccessToken()
		if err != nil {
			return nil, err
		}

		client.SetAppAccessToken(token.Data.AccessToken)
	}

	return client, err
}
