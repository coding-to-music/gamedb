package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"path"

	"github.com/gamedb/gamedb/pkg/backend/generated"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var version string
var commits string

//go:generate bash ./scripts/generate.sh

func main() {

	config.Init(version, commits, helpers.GetIP())
	log.InitZap(log.LogNameBackend)

	base := config.Config.GRPCKeysPath.Get()

	// Load the certificates from disk
	certificate, err := tls.LoadX509KeyPair(path.Join(base, "server.crt"), path.Join(base, "server.key"))
	if err != nil {
		zap.S().Errorf("could not load server key pair: %s", err)
		return
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(path.Join(base, "root.crt"))
	if err != nil {
		zap.S().Errorf("could not read ca certificate: %s", err)
		return
	}

	// Append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		zap.S().Error("failed to append client certs")
		return
	}

	// Create the channel to listen on
	lis, err := net.Listen("tcp", config.Config.BackendHostPort.Get())
	if err != nil {
		zap.S().Error(err)
		return
	}

	// Create the TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	})

	// Serve
	grpcServer := grpc.NewServer(grpc.Creds(creds))

	generated.RegisterAppsServiceServer(grpcServer, AppsServer{})
	generated.RegisterPlayersServiceServer(grpcServer, PlayersServer{})
	generated.RegisterGitHubServiceServer(grpcServer, GithubServer{})

	zap.L().Info("Starting Backend on tcp://" + config.Config.BackendHostPort.Get())

	err = grpcServer.Serve(lis)
	if err != nil {
		zap.S().Fatal(err)
	}
}
