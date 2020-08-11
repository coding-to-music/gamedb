package protos

import (
	"context"
	"sync"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
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
		conn, err = grpc.Dial(config.BackendPort(), grpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}
		ctx = context.TODO()

		log.Info("a")
	}

	return conn, ctx, nil
}
