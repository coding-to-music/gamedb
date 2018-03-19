package web

import (
	"context"
	"net/http"
	"os"

	"github.com/Jleagle/go-helpers/logger"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func CommitsHandler(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: os.Getenv("STEAM_GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	options := github.CommitsListOptions{
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 20,
		},
	}

	commits, _, err := client.Repositories.ListCommits(ctx, "steam-authority", "steam-authority", &options)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	template := commitsTemplate{}
	template.Fill(r, "Commits")
	template.Commits = commits

	returnTemplate(w, r, "commits", template)
}

type commitsTemplate struct {
	GlobalTemplate
	Commits []*github.RepositoryCommit
}
