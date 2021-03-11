package oauth

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type githubProvider struct {
}

func (c githubProvider) GetName() string {
	return "GitHub"
}

func (c githubProvider) GetIcon() string {
	return "fab fa-github"
}

func (c githubProvider) GetColour() string {
	return "#4078c0"
}

func (c githubProvider) GetEnum() ProviderEnum {
	return ProviderGithub
}

func (c githubProvider) GetType() ProviderType {
	return TypeOAuth
}

func (c githubProvider) HasEmail() bool {
	return true
}

func (c githubProvider) Redirect(w http.ResponseWriter, r *http.Request, state string) {
	conf := c.GetConfig()
	http.Redirect(w, r, conf.AuthCodeURL(state), http.StatusFound)
}

func (c githubProvider) GetUser(token *oauth2.Token) (user User, err error) {

	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+token.AccessToken)

	b, _, err := helpers.Get("https://api.github.com/user", 0, headers)
	if err != nil {
		return user, err
	}

	resp := GithubUser{}
	err = json.Unmarshal(b, &resp)
	if err != nil {
		return user, err
	}

	b, err = json.Marshal(token)
	if err != nil {
		log.ErrS(err)
	}

	user.Token = string(b)
	user.ID = strconv.FormatInt(resp.ID, 10)
	user.Username = resp.Name
	user.Email = resp.Email
	user.Avatar = resp.AvatarURL

	return user, nil
}

func (c githubProvider) GetConfig() oauth2.Config {

	return oauth2.Config{
		ClientID:     config.C.GitHubClient,
		ClientSecret: config.C.GitHubSecret,
		Scopes:       []string{""},
		RedirectURL:  config.C.GlobalSteamDomain + "/oauth/in/" + string(c.GetEnum()),
		Endpoint:     github.Endpoint,
	}
}

type GithubUser struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	ID        int64  `json:"id"`
}
