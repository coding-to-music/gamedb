package main

import (
	"fmt"
	"net"

	"github.com/gamedb/gamedb/cmd/backend/services"
	"github.com/gamedb/gamedb/pkg/protos"
	"google.golang.org/grpc"
)

var version string
var commits string

func main() {

	lis, err := net.Listen("tcp", ":"+protos.GRPCPort)
	if err != nil {
		fmt.Println(err)
		return
	}

	grpcServer := grpc.NewServer()
	protos.RegisterAppsServiceServer(grpcServer, services.AppsServer{})
	protos.RegisterGitHubServiceServer(grpcServer, services.GithubServer{})

	fmt.Println("Starting backend GRPC server")
	err = grpcServer.Serve(lis)
	fmt.Println(err)
}
