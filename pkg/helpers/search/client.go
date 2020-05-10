package search

import (
	"context"
	"sync"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/olivere/elastic/v7"
)

var (
	client *elastic.Client
	ctx    context.Context
	lock   sync.Mutex
)

type SearchResult struct {
	Keywords []string
	Name     string
	ID       uint64
	Icon     string
}

func GetElastic() (*elastic.Client, context.Context, error) {

	lock.Lock()
	defer lock.Unlock()

	var err error

	if client == nil {

		ctx = context.Background()
		client, err = elastic.NewClient(
			elastic.SetURL(config.Config.ElasticAddress.Get()),
			elastic.SetSniff(false),
			// elastic.SetHealthcheck(false),
			elastic.SetBasicAuth(config.Config.ElasticUsername.Get(), config.Config.ElasticPassword.Get()),
		)
	}

	return client, ctx, err
}
