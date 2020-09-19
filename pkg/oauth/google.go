package oauth

import (
	"encoding/json"
	"net/http"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
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
	return "4285F4"
}

func (c googleProvider) GetEnum() ProviderEnum {
	return ProviderGoogle
}

func (c googleProvider) GetConfig() oauth2.Config {

	return oauth2.Config{
		ClientID:     config.C.GoogleOauthClientID,
		ClientSecret: config.C.GoogleOauthClientSecret,
		Scopes:       []string{"profile"},
		RedirectURL:  config.C.GameDBDomain + "/oauth/in/" + string(c.GetEnum()),
		Endpoint:     google.Endpoint,
	}
}

func (c googleProvider) GetUser(_ *http.Request, token *oauth2.Token) (user User, err error) {

	body, _, err := helpers.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token="+token.AccessToken, 0, nil)
	if err != nil {
		return user, OauthError{err, "Invalid token"}
	}

	userInfo := struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
		Picture    string `json:"picture"`
		Locale     string `json:"locale"`
	}{}

	err = json.Unmarshal(body, &userInfo)
	if err != nil {
		return user, OauthError{err, "An error occurred (1005)"}
	}

	user.Token = token.AccessToken
	user.ID = userInfo.ID
	user.Username = userInfo.GivenName + " " + userInfo.FamilyName
	// user.Email = userInfo.Email // todo
	user.Avatar = userInfo.Picture

	return user, nil
}
