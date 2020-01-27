package connections

import (
	"context"
	"net/http"

	"github.com/Jleagle/session-go/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/mxpv/patreon-go"
	"golang.org/x/oauth2"
)

type patreonConnection struct {
}

func (p patreonConnection) getID(r *http.Request, token *oauth2.Token) interface{} {

	// Get Patreon user
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token.AccessToken})
	tc := oauth2.NewClient(context.TODO(), ts)

	patreonUser, err := patreon.NewClient(tc).FetchUser()
	if err != nil {
		log.Err(err)
		err = session.SetFlash(r, helpers.SessionBad, "An error occurred (1003)")
		log.Err(err)
		return nil
	}

	if !patreonUser.Data.Attributes.IsEmailVerified {
		err = session.SetFlash(r, helpers.SessionBad, "This Patreon account has not been verified")
		log.Err(err)
		return nil
	}

	return patreonUser.Data.ID
}

func (p patreonConnection) getName() string {
	return "Patreon"
}

func (p patreonConnection) getEnum() connectionEnum {
	return ConnectionPatreon
}

func (p patreonConnection) getConfig(login bool) oauth2.Config {

	var redirectURL string
	if login {
		redirectURL = config.Config.GameDBDomain.Get() + "/login/patreon-callback"
	} else {
		redirectURL = config.Config.GameDBDomain.Get() + "/settings/patreon-callback"
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

func (p patreonConnection) LinkHandler(w http.ResponseWriter, r *http.Request) {

	linkOAuth(w, r, p, false)
}

func (p patreonConnection) UnlinkHandler(w http.ResponseWriter, r *http.Request) {

	unlink(w, r, p, mongo.EventLinkPatreon)
}

func (p patreonConnection) LinkCallbackHandler(w http.ResponseWriter, r *http.Request) {

	callbackOAuth(r, p, mongo.EventUnlinkPatreon, false)

	err := session.Save(w, r)
	log.Err(err)

	http.Redirect(w, r, "/settings", http.StatusFound)
}

func (p patreonConnection) LoginHandler(w http.ResponseWriter, r *http.Request) {

	linkOAuth(w, r, p, true)
}

func (p patreonConnection) LoginCallbackHandler(w http.ResponseWriter, r *http.Request) {

	callbackOAuth(r, p, mongo.EventLogin, true)

	http.Redirect(w, r, "/login", http.StatusFound)
}
