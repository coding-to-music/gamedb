package social

import (
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/gamedb/website/pkg"
)

var twitterClient *twitter.Client

func GetTwitter() *twitter.Client {

	if twitterClient == nil {

		confi := oauth1.NewConfig(config.Config.TwitterConsumerKey.Get(), config.Config.TwitterConsumerSecret.Get())
		token := oauth1.NewToken(config.Config.TwitterAccessToken.Get(), config.Config.TwitterAccessTokenSecret.Get())

		httpClient := confi.Client(oauth1.NoContext, token)

		twitterClient = twitter.NewClient(httpClient)
	}

	return twitterClient
}
