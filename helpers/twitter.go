package helpers

import (
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/gamedb/website/config"
)

var twitterClient *twitter.Client

func GetTwitter() *twitter.Client {

	if twitterClient == nil {

		confi := oauth1.NewConfig(config.Config.TwitterConsumerKey, config.Config.TwitterConsumerSecret)
		token := oauth1.NewToken(config.Config.TwitterAccessToken, config.Config.TwitterAccessTokenSecret)

		httpClient := confi.Client(oauth1.NoContext, token)

		twitterClient = twitter.NewClient(httpClient)
	}

	return twitterClient
}
