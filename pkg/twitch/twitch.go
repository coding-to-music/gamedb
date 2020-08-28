package twitch

import (
	"errors"
	"sync"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/nicklaw5/helix"
)

var ErrNoClient = errors.New("missing twitch client env vars")

var client *helix.Client
var lock sync.Mutex

func GetTwitch() (*helix.Client, error) {

	lock.Lock()
	defer lock.Unlock()

	if config.C.TwitchClientID == "" || config.C.TwitchClientSecret == "" {
		return nil, ErrNoClient
	}

	var err error

	if client == nil {

		ops := &helix.Options{
			ClientID:     config.C.TwitchClientID,
			ClientSecret: config.C.TwitchClientSecret,
		}

		client, err = helix.NewClient(ops)
		if err != nil {
			return nil, err
		}

		token, err := client.GetAppAccessToken(nil)
		if err != nil {
			return nil, err
		}

		client.SetAppAccessToken(token.Data.AccessToken)
	}

	return client, err
}
