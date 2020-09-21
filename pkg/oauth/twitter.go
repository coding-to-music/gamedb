package oauth

import (
	"context"
	"net/http"

	twitter2 "github.com/dghubble/go-twitter/twitter"
	"github.com/gamedb/gamedb/pkg/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type twitterProvider struct {
}

func (c twitterProvider) GetName() string {
	return "Twitter"
}

func (c twitterProvider) GetIcon() string {
	return "fab fa-twitter"
}

func (c twitterProvider) GetColour() string {
	return "#1DA1F2"
}

func (c twitterProvider) GetEnum() ProviderEnum {
	return ProviderTwitter
}

func (c twitterProvider) GetConfig() oauth2.Config {

	return oauth2.Config{
		ClientID:     config.C.TwitterConsumerKey,
		ClientSecret: config.C.TwitterConsumerSecret,
		Scopes:       []string{"identity", "identity[email]"}, // identity[email] scope is only needed as the Patreon package we are using only handles v1 API
		RedirectURL:  config.C.GameDBDomain + "/oauth/in/" + string(c.GetEnum()),
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://api.twitter.com/oauth/authenticate",
			TokenURL: "https://api.twitter.com/oauth2/token",
		},
	}
}

func (c twitterProvider) GetUser(_ *http.Request, token *oauth2.Token) (user User, err error) {

	configx := &clientcredentials.Config{
		ClientID:     "consumerKey",
		ClientSecret: "consumerSecret",
		TokenURL:     "https://api.twitter.com/oauth2/token",
	}

	httpClient := configx.Client(context.Background())

	client := twitter2.NewClient(httpClient)

	t := true
	params := twitter2.AccountVerifyParams{
		IncludeEntities: &t,
		SkipStatus:      &t,
		IncludeEmail:    &t,
	}

	resp, _, err := client.Accounts.VerifyCredentials(&params)
	if err != nil {
		return user, err
	}

	user.Token = token.AccessToken
	user.ID = resp.IDStr
	user.Username = resp.ScreenName
	user.Email = resp.Email
	user.Avatar = resp.ProfileImageURL

	return user, nil
}
