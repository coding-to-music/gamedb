package connections

import (
	"context"
	"net/http"

	"github.com/Jleagle/session-go/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	pat "github.com/mxpv/patreon-go"
	"golang.org/x/oauth2"
)

type patreon struct {
}

func (p patreon) getID(r *http.Request, token *oauth2.Token) interface{} {

	// Get Patreon user
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token.AccessToken})
	tc := oauth2.NewClient(context.TODO(), ts)

	patreonUser, err := pat.NewClient(tc).FetchUser()
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

func (p patreon) getName() string {
	return "Patreon"
}

func (p patreon) getEnum() connectionEnum {
	return ConnectionPatreon
}

func (p patreon) getConfig() oauth2.Config {
	return oauth2.Config{
		ClientID:     config.Config.PatreonClientID.Get(),
		ClientSecret: config.Config.PatreonClientSecret.Get(),
		Scopes:       []string{"identity", "identity[email]"}, // identity[email] scope is only needed as the Patreon package we are using only handles v1 API
		RedirectURL:  config.Config.GameDBDomain.Get() + "/settings/patreon-callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  pat.AuthorizationURL,
			TokenURL: pat.AccessTokenURL,
		},
	}
}

func (p patreon) getEmptyVal() interface{} {
	return ""
}

func (p patreon) LinkHandler(w http.ResponseWriter, r *http.Request) {
	linkOAuth(w, r, p)
}

func (p patreon) UnlinkHandler(w http.ResponseWriter, r *http.Request) {
	unlink(w, r, p, mongo.EventLinkPatreon)
}

func (p patreon) CallbackHandler(w http.ResponseWriter, r *http.Request) {

	callbackOAuth(w, r, p, mongo.EventUnlinkPatreon)

	err := session.Save(w, r)
	log.Err(err)

	http.Redirect(w, r, "/settings", http.StatusFound)
}
