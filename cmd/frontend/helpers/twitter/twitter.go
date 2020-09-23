package twitter

import (
	"sync"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/gamedb/gamedb/pkg/config"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	client *twitter.Client
	lock   sync.Mutex
)

func GetTwitter() *twitter.Client {

	lock.Lock()
	defer lock.Unlock()

	if client == nil {

		creds := &clientcredentials.Config{
			ClientID:     config.C.TwitterConsumerKey,
			ClientSecret: config.C.TwitterConsumerSecret,
			TokenURL:     "https://api.twitter.com/oauth2/token",
		}

		httpClient := creds.Client(oauth1.NoContext)

		client = twitter.NewClient(httpClient)
	}

	return client
}
