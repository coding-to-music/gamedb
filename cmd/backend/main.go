package main

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"path/filepath"

	"github.com/gamedb/gamedb/pkg/backend/generated"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

//go:generate bash ./scripts/generate.sh

func main() {

	err := config.Init(helpers.GetIP())
	log.InitZap(log.LogNameBackend)
	defer log.Flush()
	if err != nil {
		log.ErrS(err)
		return
	}

	if config.IsProd() {
		go mongo.EnsureIndexes()
	}

	if config.C.GRPCKeysPath == "" {
		log.Err("Missing environment variables")
		return
	}

	// Load the certificates from disk
	certificate, err := tls.LoadX509KeyPair(filepath.Join(config.C.GRPCKeysPath, "server.crt"), filepath.Join(config.C.GRPCKeysPath, "server.key"))
	if err != nil {
		zap.S().Errorf("could not load server key pair: %s", err)
		return
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(filepath.Join(config.C.GRPCKeysPath, "root.crt"))
	if err != nil {
		zap.S().Errorf("could not read ca certificate: %s", err)
		return
	}

	// Append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Err("failed to append client certs")
		return
	}

	// Create the channel to listen on
	lis, err := net.Listen("tcp", config.C.BackendHostPort)
	if err != nil {
		log.ErrS(err)
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
	generated.RegisterStatsServiceServer(grpcServer, StatsServer{})
	generated.RegisterGroupsServiceServer(grpcServer, GroupsServer{})
	generated.RegisterArticlesServiceServer(grpcServer, ArticlesServer{})
	generated.RegisterPackagesServiceServer(grpcServer, PackagesServer{})

	log.Info("Starting Backend on tcp://" + lis.Addr().String())

	go func() {
		err = grpcServer.Serve(lis)
		if err != nil {
			log.ErrS(err)
		}
	}()

	helpers.KeepAlive(
		mysql.Close,
		mongo.Close,
		memcache.Close,
	)
}
