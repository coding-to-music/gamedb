package web

import (
	"context"
	"net/http"

	"github.com/gamedb/website/config"
	"github.com/gamedb/website/log"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var (
	githubContext = context.Background()
	githubClient  *github.Client
)

func init() {

	githubClient = github.NewClient(oauth2.NewClient(
		githubContext,
		oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: config.Config.GithubToken},
		)))
}

func commitsHandler(w http.ResponseWriter, r *http.Request) {

	options := github.CommitsListOptions{
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 20,
		},
	}

	t := commitsTemplate{}
	t.Fill(w, r, "Commits", "The latest commits to the Game DB code base. If a commit is on this list it does not mean it's in the latest deployment.")

	var err error
	t.Commits, _, err = githubClient.Repositories.ListCommits(githubContext, "gamedb", "website", &options)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the commits.", Error: err})
		return
	}

	err = returnTemplate(w, r, "commits", t)
	log.Err(err, r)
}

type commitsTemplate struct {
	GlobalTemplate
	Commits []*github.RepositoryCommit
}
