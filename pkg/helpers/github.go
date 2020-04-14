package helpers

import (
	"context"
	"sync"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
)

var (
	client *github.Client
	ctx    = context.Background()
	lock   sync.Mutex
)

func GetGithub() (*github.Client, context.Context) {

	lock.Lock()
	defer lock.Unlock()

	if client == nil {
		client = github.NewClient(oauth2.NewClient(
			ctx,
			oauth2.StaticTokenSource(
				&oauth2.Token{
					AccessToken: config.Config.GithubToken.Get()},
			)))
	}

	return client, ctx
}
