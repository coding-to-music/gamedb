package github

import (
	"context"
	"sync"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

var (
	client *github.Client
	ctx    = context.Background()
	lock   sync.Mutex
)

func Client() (*github.Client, context.Context) {

	lock.Lock()
	defer lock.Unlock()

	if client == nil {
		client = github.NewClient(oauth2.NewClient(
			ctx,
			oauth2.StaticTokenSource(
				&oauth2.Token{
					AccessToken: config.C.GithubToken},
			),
		))
	}

	return client, ctx
}
