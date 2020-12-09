package main

import (
	"context"

	"github.com/gamedb/gamedb/pkg/backend/generated"
)

type PackagesServer struct {
	generated.UnimplementedPackagesServiceServer
}

func (p PackagesServer) List(ctx context.Context, request *generated.ListPackagesRequest) (*generated.PackagesResponse, error) {
	panic("implement me")
}
