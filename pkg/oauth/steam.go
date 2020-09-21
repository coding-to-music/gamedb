package oauth

import (
	"net/http"
	"net/url"
	"path"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/yohcop/openid-go"
	"golang.org/x/oauth2"
)

type steamProvider struct {
}

func (c steamProvider) GetName() string {
	return "Steam"
}

func (c steamProvider) GetIcon() string {
	return "fab fa-steam"
}

func (c steamProvider) GetColour() string {
	return "#000000"
}

func (c steamProvider) GetEnum() ProviderEnum {
	return ProviderSteam
}

func (c steamProvider) Redirect(w http.ResponseWriter, r *http.Request, state string) {

	q := url.Values{}
	q.Set("openid.identity", "http://specs.openid.net/auth/2.0/identifier_select")
	q.Set("openid.claimed_id", "http://specs.openid.net/auth/2.0/identifier_select")
	q.Set("openid.ns", "http://specs.openid.net/auth/2.0")
	q.Set("openid.mode", "checkid_setup")
	q.Set("openid.realm", config.C.GameDBDomain+"/")
	q.Set("openid.return_to", config.C.GameDBDomain+"/oauth/in/steam")

	u := "https://steamcommunity.com/openid/login?" + q.Encode()

	http.Redirect(w, r, u, http.StatusFound)
}

func (c steamProvider) GetUser(r *http.Request, _ *oauth2.Token) (user User, err error) {

	// Get Steam ID
	resp, err := openid.Verify(config.C.GameDBDomain+r.URL.String(), openid.NewSimpleDiscoveryCache(), openid.NewSimpleNonceStore())
	if err != nil {
		return user, OauthError{err, "We could not verify your Steam account"}
	}

	user.ID = path.Base(resp)

	return user, nil
}
