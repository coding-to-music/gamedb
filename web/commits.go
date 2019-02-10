package web

import (
	"net/http"

	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/google/go-github/github"
)

func commitsHandler(w http.ResponseWriter, r *http.Request) {

	t := commitsTemplate{}
	t.Fill(w, r, "Commits", "The latest commits to the Game DB code base. If a commit is on this list it does not mean it's in the latest deployment.")

	client, ctx := helpers.GetGithub()

	options := github.CommitsListOptions{
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 20,
		},
	}

	var err error
	t.Commits, _, err = client.Repositories.ListCommits(ctx, "gamedb", "website", &options)
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
