package main

import (
	"github.com/gamedb/gamedb/pkg/backend/generated"
	githubHelper "github.com/gamedb/gamedb/pkg/github"
	"github.com/google/go-github/v32/github"
)

type GithubServer struct {
}

func (g GithubServer) Commits(in *generated.CommitsRequest, server generated.GitHubService_CommitsServer) error {

	client, ctx := githubHelper.GetGithub()

	ops := &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			Page:    int(in.GetPage()),
			PerPage: int(in.GetLimit()),
		},
	}
	commits, _, err := client.Repositories.ListCommits(ctx, "gamedb", "website", ops)
	if err != nil {
		return err
	}

	for _, commit := range commits {

		message := &generated.CommitResponse{
			Message: commit.GetCommit().GetMessage(),
			Time:    commit.GetCommit().GetAuthor().GetDate().Unix(),
			Link:    commit.GetHTMLURL(),
			Hash:    commit.GetSHA(),
		}
		err = server.Send(message)
		if err != nil {
			return err
		}
	}

	return nil
}
