package oauth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	sessionHelpers "github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type googleConnection struct {
	baseConnection
}

func (c googleConnection) getID(r *http.Request, token *oauth2.Token) (string, error) {

	response, err := helpers.GetWithTimeout("https://www.googleapis.com/oauth2/v2/userinfo?access_token="+token.AccessToken, 0)
	if err != nil {
		return "", oauthError{err, "Invalid token"}
	}
	defer func() {
		err := response.Body.Close()
		log.Err(err, r)
	}()

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", oauthError{err, "An error occurred (1004)"}
	}

	userInfo := struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
		Picture    string `json:"picture"`
		Locale     string `json:"locale"`
	}{}

	err = json.Unmarshal(b, &userInfo)
	if err != nil {
		return "", oauthError{err, "An error occurred (1005)"}
	}

	return userInfo.ID, nil
}

func (c googleConnection) getName() string {
	return "Google"
}

func (c googleConnection) getEnum() ConnectionEnum {
	return ConnectionGoogle
}

func (c googleConnection) getConfig(login bool) oauth2.Config {

	var redirectURL string
	if login {
		redirectURL = config.Config.GameDBDomain.Get() + "/login/oauth-callback/google"
	} else {
		redirectURL = config.Config.GameDBDomain.Get() + "/settings/oauth-callback/google"
	}

	return oauth2.Config{
		ClientID:     config.Config.GoogleOauthClientID.Get(),
		ClientSecret: config.Config.GoogleOauthClientSecret.Get(),
		Scopes:       []string{"profile"},
		RedirectURL:  redirectURL,
		Endpoint:     google.Endpoint,
	}
}

func (c googleConnection) LinkHandler(w http.ResponseWriter, r *http.Request) {

	c.linkOAuth(w, r, c, false)
}

func (c googleConnection) UnlinkHandler(w http.ResponseWriter, r *http.Request) {

	c.unlink(w, r, c, mongo.EventUnlinkGoogle)
}

func (c googleConnection) LinkCallbackHandler(w http.ResponseWriter, r *http.Request) {

	c.callbackOAuth(r, c, mongo.EventLinkGoogle, false)

	sessionHelpers.Save(w, r)

	http.Redirect(w, r, "/settings", http.StatusFound)
}

func (c googleConnection) LoginHandler(w http.ResponseWriter, r *http.Request) {

	c.linkOAuth(w, r, c, true)
}

func (c googleConnection) LoginCallbackHandler(w http.ResponseWriter, r *http.Request) {

	c.callbackOAuth(r, c, mongo.EventLogin, true)

	http.Redirect(w, r, "/login", http.StatusFound)
}
