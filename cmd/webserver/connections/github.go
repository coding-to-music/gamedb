package connections

import (
	"context"
	"net/http"

	"github.com/Jleagle/session-go/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	gh "github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type githubConnection struct {
}

func (g githubConnection) getID(r *http.Request, token *oauth2.Token) interface{} {

	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: token.AccessToken,
		},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := gh.NewClient(tc)

	// list all repositories for the authenticated user
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		log.Err(err)
		return nil
	}

	return user.GetID()
}

func (g githubConnection) getName() string {
	return "GitHub"
}

func (g githubConnection) getEnum() connectionEnum {
	return ConnectionGithub
}

func (g githubConnection) getConfig(login bool) oauth2.Config {

	var redirectURL string
	if login {
		redirectURL = config.Config.GameDBDomain.Get() + "/login/github-callback"
	} else {
		redirectURL = config.Config.GameDBDomain.Get() + "/settings/github-callback"
	}

	return oauth2.Config{
		ClientID:     config.Config.GitHubClient.Get(),
		ClientSecret: config.Config.GitHubSecret.Get(),
		Scopes:       []string{""},
		RedirectURL:  redirectURL,
		Endpoint:     github.Endpoint,
	}
}

func (g githubConnection) LinkHandler(w http.ResponseWriter, r *http.Request) {

	linkOAuth(w, r, g, false)
}

func (g githubConnection) UnlinkHandler(w http.ResponseWriter, r *http.Request) {

	unlink(w, r, g, mongo.EventUnlinkGitHub)
}

func (g githubConnection) LinkCallbackHandler(w http.ResponseWriter, r *http.Request) {

	callbackOAuth(r, g, mongo.EventLinkGitHub, false)

	err := session.Save(w, r)
	log.Err(err)

	http.Redirect(w, r, "/settings", http.StatusFound)
}

func (g githubConnection) LoginHandler(w http.ResponseWriter, r *http.Request) {

	linkOAuth(w, r, g, true)
}

func (g githubConnection) LoginCallbackHandler(w http.ResponseWriter, r *http.Request) {

	callbackOAuth(r, g, mongo.EventLogin, true)

	http.Redirect(w, r, "/login", http.StatusFound)
}
