package main

import (
	"context"

	"github.com/gamedb/gamedb/pkg/backend/generated"
)

type PackagesServer struct {
	generated.UnimplementedPackagesServiceServer
}

func (p PackagesServer) List(_ context.Context, request *generated.ListPackagesRequest) (*generated.PackagesResponse, error) {
	panic("implement me")
}
