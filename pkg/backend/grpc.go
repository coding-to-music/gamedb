package backend

import (
	"context"
	"sync"

	"github.com/gamedb/gamedb/pkg/config"
	"google.golang.org/grpc"
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

		var err error
		conn, err = grpc.Dial(config.Config.BackendClientPort.Get(), grpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}
		ctx = context.TODO()
	}

	return conn, ctx, nil
}
