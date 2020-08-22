package main

import (
	"context"

	"github.com/gamedb/gamedb/pkg/backend/generated"
)

func (p PlayersServer) Search(ctx context.Context, request *generated.SearchPlayersRequest) (*generated.PlayersElasticResponse, error) {
	panic("implement me")
}
