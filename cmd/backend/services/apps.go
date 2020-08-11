package services

import (
	"github.com/gamedb/gamedb/pkg/protos"
)

type AppsServer struct {
	protos.AppsServiceServer
}

func (a AppsServer) Search(in *protos.SearchAppsRequest, out protos.AppsService_SearchServer) error {
	return nil
}

func (a AppsServer) Apps(in *protos.ListAppsRequest, out protos.AppsService_AppsServer) error {
	return nil
}
