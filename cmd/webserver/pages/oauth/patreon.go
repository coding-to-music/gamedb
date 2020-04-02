package oauth

import (
	"context"
	"net/http"

	sessionHelpers "github.com/gamedb/gamedb/cmd/webserver/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/mxpv/patreon-go"
	"golang.org/x/oauth2"
)

type patreonConnection struct {
	baseConnection
}

func (c patreonConnection) getID(r *http.Request, token *oauth2.Token) (string, error) {

	// Get Patreon user
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token.AccessToken})
	tc := oauth2.NewClient(context.TODO(), ts)

	patreonUser, err := patreon.NewClient(tc).FetchUser()
	if err != nil {
		return "", oauthError{err, "An error occurred (1003)"}
	}

	if !patreonUser.Data.Attributes.IsEmailVerified {
		return "", oauthError{nil, "This Patreon account has not been verified"}
	}

	return patreonUser.Data.ID, nil
}

func (c patreonConnection) getName() string {
	return "Patreon"
}

func (c patreonConnection) getEnum() ConnectionEnum {
	return ConnectionPatreon
}

func (c patreonConnection) getConfig(login bool) oauth2.Config {

	var redirectURL string
	if login {
		redirectURL = config.Config.GameDBDomain.Get() + "/login/oauth-callback/patreon"
	} else {
		redirectURL = config.Config.GameDBDomain.Get() + "/settings/oauth-callback/patreon"
	}

	return oauth2.Config{
		ClientID:     config.Config.PatreonClientID.Get(),
		ClientSecret: config.Config.PatreonClientSecret.Get(),
		Scopes:       []string{"identity", "identity[email]"}, // identity[email] scope is only needed as the Patreon package we are using only handles v1 API
		RedirectURL:  redirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  patreon.AuthorizationURL,
			TokenURL: patreon.AccessTokenURL,
		},
	}
}

func (c patreonConnection) LinkHandler(w http.ResponseWriter, r *http.Request) {

	c.linkOAuth(w, r, c, false)
}

func (c patreonConnection) UnlinkHandler(w http.ResponseWriter, r *http.Request) {

	c.unlink(w, r, c, mongo.EventLinkPatreon)
}

func (c patreonConnection) LinkCallbackHandler(w http.ResponseWriter, r *http.Request) {

	c.callbackOAuth(r, c, mongo.EventUnlinkPatreon, false)

	sessionHelpers.Save(w, r)

	http.Redirect(w, r, "/settings", http.StatusFound)
}

func (c patreonConnection) LoginHandler(w http.ResponseWriter, r *http.Request) {

	c.linkOAuth(w, r, c, true)
}

func (c patreonConnection) LoginCallbackHandler(w http.ResponseWriter, r *http.Request) {

	c.callbackOAuth(r, c, mongo.EventLogin, true)

	http.Redirect(w, r, "/login", http.StatusFound)
}
