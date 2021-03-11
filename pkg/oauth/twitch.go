package oauth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/twitch"
)

type twitchProvider struct {
}

func (c twitchProvider) GetName() string {
	return "Twitch"
}

func (c twitchProvider) GetIcon() string {
	return "fab fa-twitch"
}

func (c twitchProvider) GetColour() string {
	return "#6441A4"
}

func (c twitchProvider) GetEnum() ProviderEnum {
	return ProviderTwitch
}

func (c twitchProvider) GetType() ProviderType {
	return TypeOAuth
}

func (c twitchProvider) HasEmail() bool {
	return true
}

func (c twitchProvider) Redirect(w http.ResponseWriter, r *http.Request, state string) {
	conf := c.GetConfig()
	http.Redirect(w, r, conf.AuthCodeURL(state), http.StatusFound)
}

func (c twitchProvider) GetUser(token *oauth2.Token) (user User, err error) {

	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+token.AccessToken)
	headers.Add("Client-Id", config.C.TwitchClientID)

	b, _, err := helpers.Get("https://api.twitch.tv/helix/users", 0, headers)
	if err != nil {
		return user, err
	}

	resp := TwitchUser{}
	err = json.Unmarshal(b, &resp)
	if err != nil {
		return user, err
	}

	if len(resp.Data) == 0 {
		return user, errors.New("no user returned from api")
	}

	b, err = json.Marshal(token)
	if err != nil {
		log.ErrS(err)
	}

	user.Token = string(b)
	user.ID = resp.Data[0].ID
	user.Username = resp.Data[0].DisplayName
	user.Email = resp.Data[0].Email
	user.Avatar = resp.Data[0].ProfileImageURL

	return user, nil
}

func (c twitchProvider) GetConfig() oauth2.Config {

	return oauth2.Config{
		ClientID:     config.C.TwitchClientID,
		ClientSecret: config.C.TwitchClientSecret,
		Scopes:       []string{"user:read:email"},
		RedirectURL:  config.C.GlobalSteamDomain + "/oauth/in/" + string(c.GetEnum()),
		Endpoint:     twitch.Endpoint,
	}
}

type TwitchUser struct {
	Data []struct {
		ID              string `json:"id"`
		Login           string `json:"login"`
		DisplayName     string `json:"display_name"`
		Type            string `json:"type"`
		BroadcasterType string `json:"broadcaster_type"`
		Description     string `json:"description"`
		ProfileImageURL string `json:"profile_image_url"`
		OfflineImageURL string `json:"offline_image_url"`
		ViewCount       int    `json:"view_count"`
		Email           string `json:"email"`
	} `json:"data"`
}
