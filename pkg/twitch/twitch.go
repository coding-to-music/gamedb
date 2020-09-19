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

		if config.C.TwitchClientID == "" || config.C.TwitchClientSecret == "" {
			return nil, config.ErrMissingEnvironmentVariable
		}

		ops := &helix.Options{
			ClientID:     config.C.TwitchClientID,
			ClientSecret: config.C.TwitchClientSecret,
		}

		client, err = helix.NewClient(ops)
		if err != nil {
			return nil, err
		}

		token, err := client.RequestAppAccessToken(nil)
		if err != nil {
			return nil, err
		}

		client.SetAppAccessToken(token.Data.AccessToken)
	}

	return client, err
}
