package oauth

import (
	"context"
	"net/http"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/mxpv/patreon-go"
	"golang.org/x/oauth2"
)

type patreonProvider struct {
}

func (c patreonProvider) GetName() string {
	return "Patreon"
}

func (c patreonProvider) GetIcon() string {
	return "fab fa-patreon"
}

func (c patreonProvider) GetColour() string {
	return "f96854"
}

func (c patreonProvider) GetEnum() ProviderEnum {
	return ProviderPatreon
}

func (c patreonProvider) GetConfig() oauth2.Config {

	return oauth2.Config{
		ClientID:     config.C.PatreonClientID,
		ClientSecret: config.C.PatreonClientSecret,
		Scopes:       []string{"identity", "identity[email]"}, // identity[email] scope is only needed as the Patreon package we are using only handles v1 API
		RedirectURL:  config.C.GameDBDomain + "/oauth/in/" + string(c.GetEnum()),
		Endpoint: oauth2.Endpoint{
			AuthURL:  patreon.AuthorizationURL,
			TokenURL: patreon.AccessTokenURL,
		},
	}
}

func (c patreonProvider) GetUser(_ *http.Request, token *oauth2.Token) (user User, err error) {

	// Get Patreon user
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token.AccessToken})
	tc := oauth2.NewClient(context.TODO(), ts)

	resp, err := patreon.NewClient(tc).FetchUser()
	if err != nil {
		return user, OauthError{err, "An error occurred (1003)"}
	}

	// if !patreonUser.Data.Attributes.IsEmailVerified {
	// 	return "", OauthError{nil, "This Patreon account has not been verified"}
	// }

	user.Token = token.AccessToken
	user.ID = resp.Data.ID
	user.Username = resp.Data.Attributes.FullName
	user.Email = resp.Data.Attributes.Email
	user.Avatar = resp.Data.Attributes.ImageURL

	return user, nil
}
