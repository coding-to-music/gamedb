package backend

import (
	"context"
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

		creds, err := credentials.NewClientTLSFromFile(path.Join(config.Config.InfraPath.Get(), "/grpc/app.crt"), "")
		if err != nil {
			return nil, nil, err
		}

		conn, err = grpc.Dial(config.Config.BackendClientPort.Get(), grpc.WithTransportCredentials(creds))
		if err != nil {
			return nil, nil, err
		}
		ctx = context.Background()
	}

	return conn, ctx, nil
}
