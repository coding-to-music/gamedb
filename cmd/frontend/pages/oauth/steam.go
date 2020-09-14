package oauth

import (
	"net/http"
	"path"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/yohcop/openid-go"
	"golang.org/x/oauth2"
)

type steamConnection struct {
	baseConnection
}

func (c steamConnection) getID(r *http.Request, token *oauth2.Token) (string, error) {

	// Get Steam ID
	openID, err := openid.Verify(config.C.GameDBDomain+r.URL.String(), openid.NewSimpleDiscoveryCache(), openid.NewSimpleNonceStore())
	if err != nil {
		return "", oauthError{err, "We could not verify your Steam account"}
	}

	return path.Base(openID), nil
}

func (c steamConnection) getName() string {
	return "Steam"
}

func (c steamConnection) getEnum() ConnectionEnum {
	return ConnectionSteam
}

func (c steamConnection) getConfig(login bool) oauth2.Config {
	return oauth2.Config{}
}

func (c steamConnection) LinkHandler(w http.ResponseWriter, r *http.Request) {
}

func (c steamConnection) UnlinkHandler(w http.ResponseWriter, r *http.Request) {
	c.unlink(w, r, c, mongo.EventUnlinkSteam)
}

func (c steamConnection) LinkCallbackHandler(w http.ResponseWriter, r *http.Request) {

	c.callback(r, c, mongo.EventLinkSteam, nil, false)

	session.Save(w, r)

	http.Redirect(w, r, "/settings", http.StatusFound)
}

func (c steamConnection) LoginHandler(w http.ResponseWriter, r *http.Request) {

	// id := c.getID(r, nil)
}

func (c steamConnection) LoginCallbackHandler(w http.ResponseWriter, r *http.Request) {

	c.callback(r, c, mongo.EventLogin, nil, true)

	session.Save(w, r)

	http.Redirect(w, r, "/login", http.StatusFound)
}
