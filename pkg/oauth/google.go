package oauth

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type googleProvider struct {
}

func (c googleProvider) GetName() string {
	return "Google"
}

func (c googleProvider) GetIcon() string {
	return "fab fa-google"
}

func (c googleProvider) GetColour() string {
	return "#4285F4"
}

func (c googleProvider) GetEnum() ProviderEnum {
	return ProviderGoogle
}

func (c googleProvider) GetType() ProviderType {
	return TypeOAuth
}

func (c googleProvider) Redirect(w http.ResponseWriter, r *http.Request, state string) {
	conf := c.GetConfig()
	http.Redirect(w, r, conf.AuthCodeURL(state), http.StatusFound)
}

func (c googleProvider) GetUser(token *oauth2.Token) (user User, err error) {

	q := url.Values{}
	q.Set("access_token", token.AccessToken)

	body, _, err := helpers.Get("https://openidconnect.googleapis.com/v1/userinfo?"+q.Encode(), 0, nil)
	if err != nil {
		return user, OauthError{err, "Invalid token"}
	}

	userInfo := struct {
		Sub     string `json:"sub"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
		Email   string `json:"email"`
		// GivenName     string `json:"given_name"`
		// FamilyName    string `json:"family_name"`
		// EmailVerified bool   `json:"email_verified"`
		// Locale        string `json:"locale"`
	}{}

	err = json.Unmarshal(body, &userInfo)
	if err != nil {
		return user, OauthError{err, "An error occurred (1005)"}
	}

	b, err := json.Marshal(token)
	if err != nil {
		log.ErrS(err)
	}

	user.Token = string(b)
	user.ID = userInfo.Sub
	user.Username = userInfo.Name
	user.Email = userInfo.Email
	user.Avatar = userInfo.Picture

	return user, nil
}

func (c googleProvider) GetConfig() oauth2.Config {

	return oauth2.Config{
		ClientID:     config.C.GoogleOauthClientID,
		ClientSecret: config.C.GoogleOauthClientSecret,
		Scopes:       []string{"profile", "email"},
		RedirectURL:  config.C.GameDBDomain + "/oauth/in/" + string(c.GetEnum()),
		Endpoint:     google.Endpoint,
	}
}
