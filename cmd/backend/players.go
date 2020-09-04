package main

import (
	"context"

	"github.com/gamedb/gamedb/pkg/backend/generated"
)

type PlayersServer struct {
}

func (p PlayersServer) List(ctx context.Context, request *generated.ListPlayersRequest) (*generated.PlayersMongoResponse, error) {
	panic("implement me")
}
