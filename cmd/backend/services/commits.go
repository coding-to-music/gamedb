package services

import (
	githubHelper "github.com/gamedb/gamedb/pkg/github"
	"github.com/gamedb/gamedb/pkg/protos"
	"github.com/google/go-github/v28/github"
)

type GithubServer struct {
	protos.GitHubServiceServer
}

func (g GithubServer) Commits(in *protos.CommitsRequest, server protos.GitHubService_CommitsServer) error {

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

		message := &protos.CommitResponse{
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
