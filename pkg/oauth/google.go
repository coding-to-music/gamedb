package oauth

import (
	"encoding/json"
	"net/http"

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

func (c googleProvider) HasEmail() bool {
	return true
}

func (c googleProvider) Redirect(w http.ResponseWriter, r *http.Request, state string) {
	conf := c.GetConfig()
	http.Redirect(w, r, conf.AuthCodeURL(state), http.StatusFound)
}

func (c googleProvider) GetUser(token *oauth2.Token) (user User, err error) {

	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+token.AccessToken)

	b, _, err := helpers.Get("https://openidconnect.googleapis.com/v1/userinfo", 0, headers)
	if err != nil {
		return user, err
	}

	resp := GoogleUser{}
	err = json.Unmarshal(b, &resp)
	if err != nil {
		return user, err
	}

	b, err = json.Marshal(token)
	if err != nil {
		log.ErrS(err)
	}

	user.Token = string(b)
	user.ID = resp.Sub
	user.Username = resp.Name
	user.Email = resp.Email
	user.Avatar = resp.Picture

	return user, nil
}

func (c googleProvider) GetConfig() oauth2.Config {

	return oauth2.Config{
		ClientID:     config.C.GoogleOauthClientID,
		ClientSecret: config.C.GoogleOauthClientSecret,
		Scopes:       []string{"profile", "email"},
		RedirectURL:  config.C.GlobalSteamDomain + "/oauth/in/" + string(c.GetEnum()),
		Endpoint:     google.Endpoint,
	}
}

type GoogleUser struct {
	Sub     string `json:"sub"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	Email   string `json:"email"`
}
