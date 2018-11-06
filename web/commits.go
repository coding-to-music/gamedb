package web

import (
	"context"
	"net/http"

	"github.com/gamedb/website/logging"
	"github.com/google/go-github/github"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

var (
	githubContext = context.Background()
	githubClient  *github.Client
)

// Called from main
func InitCommits() {

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

	t := commitsTemplate{}
	t.Fill(w, r, "Commits")

	var err error
	t.Commits, _, err = githubClient.Repositories.ListCommits(githubContext, "gamedb", "website", &options)
	if err != nil {
		logging.Error(err)
		returnErrorTemplate(w, r, 500, "Can't connect to GitHub")
		return
	}

	returnTemplate(w, r, "commits", t)
}

type commitsTemplate struct {
	GlobalTemplate
	Commits []*github.RepositoryCommit
}
