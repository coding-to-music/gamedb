package web

import (
	"net/http"

	"github.com/gamedb/website/config"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/google/go-github/github"
)

func commitsHandler(w http.ResponseWriter, r *http.Request) {

	t := commitsTemplate{}
	t.fill(w, r, "Commits", "The last commits to the Game DB code base.")

	client, ctx := helpers.GetGithub()

	var err error
	commits, _, err := client.Repositories.ListCommits(ctx, "gamedb", "website", &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	})

	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the commits.", Error: err})
		return
	}

	var deployed bool
	for _, v := range commits {

		if v.GetSHA() == config.Config.CommitHash {
			deployed = true
		}

		t.Commits = append(t.Commits, commit{
			Message:   v.Commit.GetMessage(),
			Time:      v.Commit.Author.Date.Unix(),
			Deployed:  deployed,
			Link:      v.GetHTMLURL(),
			Highlight: v.GetSHA() == config.Config.CommitHash,
			Hash:      v.GetSHA()[0:7],
		})
	}

	err = returnTemplate(w, r, "commits", t)
	log.Err(err, r)
}

type commitsTemplate struct {
	GlobalTemplate
	Commits []commit
	Hash    string
}

type commit struct {
	Message   string
	Deployed  bool
	Time      int64
	Link      string
	Highlight bool
	Hash      string
}
