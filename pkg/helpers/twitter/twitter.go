package twitter

import (
	"sync"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/gamedb/gamedb/pkg/config"
)

var (
	client *twitter.Client
	lock   sync.Mutex
)

func GetTwitter() *twitter.Client {

	lock.Lock()
	defer lock.Unlock()

	if client == nil {

		confi := oauth1.NewConfig(config.Config.TwitterConsumerKey.Get(), config.Config.TwitterConsumerSecret.Get())
		token := oauth1.NewToken(config.Config.TwitterAccessToken.Get(), config.Config.TwitterAccessTokenSecret.Get())

		httpClient := confi.Client(oauth1.NoContext, token)

		client = twitter.NewClient(httpClient)
	}

	return client
}
