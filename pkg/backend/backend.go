package backend

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
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

	if config.C.GRPCKeysPath == "" {
		return nil, nil, errors.New("missing environment variables")
	}

	if conn == nil {

		// Load the client certificates from disk
		certificate, err := tls.LoadX509KeyPair(filepath.Join(config.C.GRPCKeysPath, "client.crt"), filepath.Join(config.C.GRPCKeysPath, "client.key"))
		if err != nil {
			return nil, nil, fmt.Errorf("could not load client key pair: %s", err)
		}

		// Create a certificate pool from the certificate authority
		certPool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(filepath.Join(config.C.GRPCKeysPath, "root.crt"))
		if err != nil {
			return nil, nil, fmt.Errorf("could not read ca certificate: %s", err)
		}

		// Append the certificates from the CA
		ok := certPool.AppendCertsFromPEM(ca)
		if !ok {
			return nil, nil, errors.New("failed to append ca certs")
		}

		creds := credentials.NewTLS(&tls.Config{
			ServerName:   "server", // Must match the key name
			Certificates: []tls.Certificate{certificate},
			RootCAs:      certPool,
		})

		// Retry
		operation := func() (err error) {

			c, err := grpc.Dial(config.C.BackendClientPort, grpc.WithTransportCredentials(creds))
			if err == nil {
				conn = c
				ctx = context.Background()
			}
			return err
		}

		policy := backoff.NewExponentialBackOff()
		policy.InitialInterval = time.Second / 2
		policy.MaxInterval = time.Second * 10

		err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Warn(err.Error()) })
		if err != nil {
			return nil, nil, err
		}
	}

	return conn, ctx, nil
}
