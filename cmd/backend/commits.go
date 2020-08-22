package main

import (
	"context"

	"github.com/gamedb/gamedb/pkg/backend/generated"
	githubHelper "github.com/gamedb/gamedb/pkg/github"
	"github.com/google/go-github/v32/github"
)

func (g GithubServer) Commits(ctx context.Context, request *generated.CommitsRequest) (response *generated.CommitsResponse, err error) {

	client, ctx := githubHelper.GetGithub()

	ops := &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			Page:    int(request.GetPagination().GetPage()),
			PerPage: int(request.GetPagination().GetLimit()),
		},
	}

	commits, _, err := client.Repositories.ListCommits(ctx, "gamedb", "website", ops)
	if err != nil {
		return response, err
	}

	response = &generated.CommitsResponse{}
	for _, commit := range commits {

		response.Commits = append(response.Commits, &generated.CommitResponse{
			Message: commit.GetCommit().GetMessage(),
			Time:    commit.GetCommit().GetAuthor().GetDate().Unix(),
			Link:    commit.GetHTMLURL(),
			Hash:    commit.GetSHA(),
		})
	}

	return response, err
}
