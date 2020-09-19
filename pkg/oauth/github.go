package oauth

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gamedb/gamedb/pkg/config"
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
	return "4078c0"
}

func (c githubProvider) GetEnum() ProviderEnum {
	return ProviderGithub
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

func (c githubProvider) GetUser(_ *http.Request, token *oauth2.Token) (user User, err error) {

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

	user.Token = token.AccessToken
	user.ID = strconv.FormatInt(resp.GetID(), 10)
	user.Username = resp.GetName()
	user.Email = resp.GetEmail()
	user.Avatar = resp.GetAvatarURL()

	return user, nil
}
