package oauth

import (
	"encoding/json"
	"net/http"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"golang.org/x/oauth2"
)

const UserAgent = "globalsteam.online"

type redditProvider struct {
}

func (c redditProvider) GetName() string {
	return "Reddit"
}

func (c redditProvider) GetIcon() string {
	return "fab fa-reddit"
}

func (c redditProvider) GetColour() string {
	return "#FF5700"
}

func (c redditProvider) GetEnum() ProviderEnum {
	return ProviderReddit
}

func (c redditProvider) GetType() ProviderType {
	return TypeOAuth
}

func (c redditProvider) Redirect(w http.ResponseWriter, r *http.Request, state string) {
	conf := c.GetConfig()
	http.Redirect(w, r, conf.AuthCodeURL(state), http.StatusFound)
}

func (c redditProvider) HasEmail() bool {
	return false
}

func (c redditProvider) GetUser(token *oauth2.Token) (user User, err error) {

	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+token.AccessToken)
	headers.Add("User-Agent", UserAgent)

	b, _, err := helpers.Get("https://oauth.reddit.com/api/v1/me", 0, headers)
	if err != nil {
		return user, err
	}

	resp := RedditUser{}
	err = json.Unmarshal(b, &resp)
	if err != nil {
		return user, err
	}

	b, err = json.Marshal(token)
	if err != nil {
		log.ErrS(err)
	}

	user.Token = string(b)
	user.ID = resp.ID
	user.Username = resp.Name
	user.Avatar = resp.IconImg

	return user, nil
}

func (c redditProvider) GetConfig() oauth2.Config {

	return oauth2.Config{
		ClientID:     config.C.RedditClient,
		ClientSecret: config.C.RedditSecret,
		Scopes:       []string{"identity"},
		RedirectURL:  config.C.GlobalSteamDomain + "/oauth/in/" + string(c.GetEnum()),
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.reddit.com/api/v1/authorize",
			TokenURL: "https://www.reddit.com/api/v1/access_token",
		},
	}
}

type RedditUser struct {
	ID      string `json:"id"`
	IconImg string `json:"icon_img"`
	Name    string `json:"name"`
}
