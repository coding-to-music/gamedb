package main

import (
	"context"

	"github.com/gamedb/gamedb/pkg/backend/generated"
)

type ArticlesServer struct {
	generated.UnimplementedArticlesServiceServer
}

func (a ArticlesServer) List(ctx context.Context, request *generated.ListArticlesRequest) (*generated.ArticlesResponse, error) {
	panic("implement me")
}
