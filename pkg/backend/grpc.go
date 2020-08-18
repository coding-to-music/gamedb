package backend

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"sync"

	"github.com/gamedb/gamedb/pkg/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	conn *grpc.ClientConn
	ctx  context.Context
	lock sync.Mutex
)

func GetClient() (*grpc.ClientConn, context.Context, error) {

	lock.Lock()
	defer lock.Unlock()

	if conn == nil {

		base := path.Join(config.Config.InfraPath.Get(), "/grpc")

		// Load the client certificates from disk
		certificate, err := tls.LoadX509KeyPair(path.Join(base, "client.crt"), path.Join(base, "client.key"))
		if err != nil {
			return nil, nil, fmt.Errorf("could not load client key pair: %s", err)
		}

		// Create a certificate pool from the certificate authority
		certPool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(path.Join(base, "ca.crt"))
		if err != nil {
			return nil, nil, fmt.Errorf("could not read ca certificate: %s", err)
		}

		// Append the certificates from the CA
		ok := certPool.AppendCertsFromPEM(ca)
		if !ok {
			return nil, nil, errors.New("failed to append ca certs")
		}

		creds := credentials.NewTLS(&tls.Config{
			ServerName:   config.Config.BackendClientPort.Get(),
			Certificates: []tls.Certificate{certificate},
			RootCAs:      certPool,
		})

		conn, err = grpc.Dial(config.Config.BackendClientPort.Get(), grpc.WithTransportCredentials(creds))
		if err != nil {
			return nil, nil, err
		}

		ctx = context.Background()
	}

	return conn, ctx, nil
}
