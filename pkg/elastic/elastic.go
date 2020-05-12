package elastic

import (
	"context"
	"sync"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/olivere/elastic/v7"
)

const (
	IndexApps    = "apps"
	IndexPlayers = "players"
)

var (
	client *elastic.Client
	ctx    context.Context
	lock   sync.Mutex
)

func GetElastic() (*elastic.Client, context.Context, error) {

	lock.Lock()
	defer lock.Unlock()

	var err error

	if client == nil {

		ctx = context.Background()

		ops := []elastic.ClientOptionFunc{
			elastic.SetURL(config.Config.ElasticAddress.Get()),
			elastic.SetSniff(false),
			elastic.SetBasicAuth(config.Config.ElasticUsername.Get(), config.Config.ElasticPassword.Get()),
		}

		if config.IsLocal() || true {
			ops = append(ops, elastic.SetHealthcheck(false))
		}

		client, err = elastic.NewClient(ops...)
	}

	return client, ctx, err
}


