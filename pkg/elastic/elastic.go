package elastic

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
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

func DeleteAndRebuildAppsIndex() {

	var priceProperties = map[string]interface{}{}
	for _, v := range steamapi.ProductCCs {
		priceProperties[string(v)] = map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"currency":         map[string]interface{}{"type": "keyword"},
				"discount_percent": map[string]interface{}{"type": "integer"},
				"final":            map[string]interface{}{"type": "integer"},
				"individual":       map[string]interface{}{"type": "integer"},
				"initial":          map[string]interface{}{"type": "integer"},
			},
		}
	}

	var mapping = map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "integer",
				},
				"name": map[string]interface{}{
					"type": "text",
				},
				"icon": map[string]interface{}{
					"enabled": false,
				},
				"players": map[string]interface{}{
					"type": "integer",
				},
				"followers": map[string]interface{}{
					"type": "integer",
				},
				"score": map[string]interface{}{
					"type": "half_float",
				},
				"prices": map[string]interface{}{
					"type":       "object",
					"properties": priceProperties,
				},
				"tags": map[string]interface{}{
					"type": "integer",
				},
				"genres": map[string]interface{}{
					"type": "integer",
				},
				"categories": map[string]interface{}{
					"type": "integer",
				},
				"publishers": map[string]interface{}{
					"type": "integer",
				},
				"developers": map[string]interface{}{
					"type": "integer",
				},
				"type": map[string]interface{}{
					"type": "keyword",
				},
				"platforms": map[string]interface{}{
					"type": "keyword",
				},
			},
		},
	}

	client, ctx, err := GetElastic()
	if err != nil {
		log.Err(err)
		return
	}

	_, err = client.DeleteIndex("apps").Do(ctx)
	if err != nil {
		log.Err(err)
		return
	}

	time.Sleep(time.Second)

	createIndex, err := client.CreateIndex(IndexApps).BodyJson(mapping).Do(ctx)
	if err != nil {
		log.Err(err)
		return
	}

	if !createIndex.Acknowledged {
		log.Err(errors.New("not acknowledged"))
	}
}
