package helpers

import (
	"context"

	"github.com/gamedb/website/config"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var (
	githubContext = context.Background()
	githubClient  *github.Client
)

func GetGithub() (*github.Client, context.Context) {

	if githubClient == nil {
		githubClient = github.NewClient(oauth2.NewClient(
			githubContext,
			oauth2.StaticTokenSource(
				&oauth2.Token{
					AccessToken: config.Config.GithubToken},
			)))
	}

	return githubClient, githubContext
}
