package main

import (
	"fmt"
	"net"

	"github.com/gamedb/gamedb/pkg/backend"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"google.golang.org/grpc"
)

var version string
var commits string

func main() {

	config.Init(version, commits, helpers.GetIP())
	log.Initialise(log.LogNameBackend)

	lis, err := net.Listen("tcp", config.Config.BackendHostPort.Get())
	if err != nil {
		fmt.Println(err)
		return
	}

	grpcServer := grpc.NewServer()
	backend.RegisterAppsServiceServer(grpcServer, AppsServer{})
	backend.RegisterPlayersServiceServer(grpcServer, PlayersServer{})
	backend.RegisterGitHubServiceServer(grpcServer, GithubServer{})

	fmt.Println("Starting Backend on " + config.Config.BackendHostPort.Get())
	err = grpcServer.Serve(lis)
	fmt.Println(err)
}
