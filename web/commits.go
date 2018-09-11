package web

import (
	"context"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/spf13/viper"
	"github.com/steam-authority/steam-authority/logger"
	"golang.org/x/oauth2"
)

var (
	githubContext = context.Background()
	githubClient  *github.Client
)

func init() {

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: viper.GetString("GITHUB_TOKEN")},
	)

	tc := oauth2.NewClient(githubContext, ts)

	githubClient = github.NewClient(tc)
}

func CommitsHandler(w http.ResponseWriter, r *http.Request) {

	options := github.CommitsListOptions{
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 20,
		},
	}

	commits, _, err := githubClient.Repositories.ListCommits(githubContext, "steam-authority", "steam-authority", &options)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, "Can't connect to GitHub")
		return
	}

	t := commitsTemplate{}
	t.Fill(w, r, "Commits")
	t.Commits = commits

	returnTemplate(w, r, "commits", t)
}

type commitsTemplate struct {
	GlobalTemplate
	Commits []*github.RepositoryCommit
}
