package main

import (
	"context"

	"github.com/gamedb/gamedb/pkg/backend"
)

type PlayersServer struct {
}

func (p PlayersServer) Search(ctx context.Context, request *backend.SearchPlayersRequest) (*backend.PlayersElasticResponse, error) {
	panic("implement me")
}

func (p PlayersServer) List(ctx context.Context, request *backend.ListPlayersRequest) (*backend.PlayersMongoResponse, error) {
	panic("implement me")
}
