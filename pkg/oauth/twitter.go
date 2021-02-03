package oauth

import (
	"encoding/json"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	twitter2 "github.com/dghubble/oauth1/twitter"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
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

func (c twitterProvider) GetType() ProviderType {
	return TypeOAuth
}

func (c twitterProvider) HasEmail() bool {
	return true
}

func (c twitterProvider) Redirect() (redirect string, secret string, err error) {

	conf := c.GetConfig()
	requestToken, requestSecret, err := conf.RequestToken()
	if err != nil {
		return "", "", err
	}

	authorizationURL, err := conf.AuthorizationURL(requestToken)
	if err != nil {
		return "", "", err
	}

	return authorizationURL.String(), requestSecret, nil
}

func (c twitterProvider) GetUser(token *oauth1.Token) (user User, err error) {

	conf := oauth1.NewConfig(config.C.TwitterConsumerKey, config.C.TwitterConsumerSecret)
	httpClient := conf.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	// Get user
	t := true
	params := twitter.AccountVerifyParams{
		IncludeEntities: &t,
		SkipStatus:      &t,
		IncludeEmail:    &t,
	}

	resp, _, err := client.Accounts.VerifyCredentials(&params)
	if err != nil {
		return user, err
	}

	b, err := json.Marshal(token)
	if err != nil {
		log.ErrS(err)
	}

	user.Token = string(b)
	user.ID = resp.IDStr
	user.Username = resp.Name
	user.Email = resp.Email
	user.Avatar = resp.ProfileImageURL

	// Add friend
	friendParams := &twitter.FriendshipCreateParams{
		ScreenName: "gamedb_online",
		UserID:     0,
		Follow:     &t,
	}

	_, _, err = client.Friendships.Create(friendParams)
	if err != nil {
		log.ErrS(err)
	}

	return user, nil
}

func (c twitterProvider) GetConfig() oauth1.Config {

	return oauth1.Config{
		ConsumerKey:    config.C.TwitterConsumerKey,
		ConsumerSecret: config.C.TwitterConsumerSecret,
		CallbackURL:    config.C.GameDBDomain + "/oauth/in/" + string(c.GetEnum()),
		Endpoint:       twitter2.AuthorizeEndpoint,
	}
}
