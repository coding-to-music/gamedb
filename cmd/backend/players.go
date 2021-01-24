package main

import (
	"context"

	"github.com/gamedb/gamedb/pkg/backend/generated"
)

type PlayersServer struct {
	generated.UnimplementedPlayersServiceServer
}

func (p PlayersServer) List(_ context.Context, request *generated.ListPlayersRequest) (*generated.PlayersMongoResponse, error) {
	panic("implement me")
}

func (p PlayersServer) Search(_ context.Context, request *generated.SearchPlayersRequest) (*generated.PlayersElasticResponse, error) {
	panic("implement me")
}
