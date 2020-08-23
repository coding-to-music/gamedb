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

		confi := oauth1.NewConfig(config.C.TwitterConsumerKey, config.C.TwitterConsumerSecret)
		token := oauth1.NewToken(config.C.TwitterAccessToken, config.C.TwitterAccessTokenSecret)

		httpClient := confi.Client(oauth1.NoContext, token)

		client = twitter.NewClient(httpClient)
	}

	return client
}
