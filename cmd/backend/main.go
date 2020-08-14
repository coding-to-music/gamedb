package main

import (
	"net"

	"github.com/gamedb/gamedb/pkg/backend"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var version string
var commits string

func main() {

	config.Init(version, commits, helpers.GetIP())
	log.InitZap(log.LogNameBackend)

	lis, err := net.Listen("tcp", config.Config.BackendHostPort.Get())
	if err != nil {
		zap.S().Error(err)
		return
	}

	grpcServer := grpc.NewServer()
	backend.RegisterAppsServiceServer(grpcServer, AppsServer{})
	backend.RegisterPlayersServiceServer(grpcServer, PlayersServer{})
	backend.RegisterGitHubServiceServer(grpcServer, GithubServer{})

	zap.S().Info("Starting Backend on " + config.Config.BackendHostPort.Get())

	err = grpcServer.Serve(lis)
	if err != nil {
		zap.S().Error(err)
	}
}
