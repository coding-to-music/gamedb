package helpers

import (
	"sync"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/gamedb/gamedb/pkg/config"
)

var (
	twitterClient *twitter.Client
	twitterLock   sync.Mutex
)

func GetTwitter() *twitter.Client {

	twitterLock.Lock()
	defer twitterLock.Unlock()

	if twitterClient == nil {

		confi := oauth1.NewConfig(config.Config.TwitterConsumerKey.Get(), config.Config.TwitterConsumerSecret.Get())
		token := oauth1.NewToken(config.Config.TwitterAccessToken.Get(), config.Config.TwitterAccessTokenSecret.Get())

		httpClient := confi.Client(oauth1.NoContext, token)

		twitterClient = twitter.NewClient(httpClient)
	}

	return twitterClient
}
