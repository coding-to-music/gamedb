package main

import (
	"fmt"
	"net"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/protos"
	"google.golang.org/grpc"
)

var version string
var commits string

func main() {

	config.Init(version, commits, helpers.GetIP())
	log.Initialise(log.LogNameBackend)

	lis, err := net.Listen("tcp", config.BackendPort())
	if err != nil {
		fmt.Println(err)
		return
	}

	grpcServer := grpc.NewServer()
	protos.RegisterAppsServiceServer(grpcServer, AppsServer{})
	protos.RegisterGitHubServiceServer(grpcServer, GithubServer{})

	fmt.Println("Starting Backend on " + config.BackendPort())
	err = grpcServer.Serve(lis)
	fmt.Println(err)
}
