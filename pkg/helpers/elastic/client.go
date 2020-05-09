package elastic

import (
	"sync"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gamedb/gamedb/pkg/config"
)

var (
	client *elasticsearch.Client
	lock   sync.Mutex
)

func GetElastic() (*elasticsearch.Client, error) {

	lock.Lock()
	defer lock.Unlock()

	var err error

	if client == nil {

		conf := elasticsearch.Config{
			Addresses: []string{config.Config.ElasticAddress.Get()},
			Username:  config.Config.ElasticUsername.Get(),
			Password:  config.Config.ElasticPassword.Get(),
		}

		client, err = elasticsearch.NewClient(conf)
	}

	return client, err
}
