package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	gh "github.com/google/go-github/v32/github"
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

func (c githubProvider) Redirect(w http.ResponseWriter, r *http.Request, state string) {
	conf := c.GetConfig()
	http.Redirect(w, r, conf.AuthCodeURL(state), http.StatusFound)
}

func (c githubProvider) GetUser(token *oauth2.Token) (user User, err error) {

	ctx := context.Background()

	client := gh.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: token.AccessToken,
		},
	)))

	resp, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return user, err
	}

	b, err := json.Marshal(token)
	if err != nil {
		log.ErrS(err)
	}

	user.Token = string(b)
	user.ID = strconv.FormatInt(resp.GetID(), 10)
	user.Username = resp.GetName()
	user.Email = resp.GetEmail()
	user.Avatar = resp.GetAvatarURL()

	return user, nil
}

func (c githubProvider) GetConfig() oauth2.Config {

	return oauth2.Config{
		ClientID:     config.C.GitHubClient,
		ClientSecret: config.C.GitHubSecret,
		Scopes:       []string{""},
		RedirectURL:  config.C.GameDBDomain + "/oauth/in/" + string(c.GetEnum()),
		Endpoint:     github.Endpoint,
	}
}
