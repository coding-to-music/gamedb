package elastic

import (
	"context"
	"errors"
	"sync"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/olivere/elastic/v7"
)

const (
	IndexAchievements = "achievements"
	IndexGroups       = "groups"
	IndexArticles     = "articles"
	IndexApps         = "apps"
	IndexPlayers      = "players"
)

var (
	settings                = map[string]interface{}{"number_of_shards": 1, "number_of_replicas": 0}
	fieldTypeInteger        = map[string]interface{}{"type": "integer"}    // int32
	fieldTypeHalfFloat      = map[string]interface{}{"type": "half_float"} // float16
	fieldTypeLong           = map[string]interface{}{"type": "long"}       // int64
	fieldTypeBool           = map[string]interface{}{"type": "boolean"}    // bool
	fieldTypeKeyword        = map[string]interface{}{"type": "keyword"}    // Exact matches
	fieldTypeText           = map[string]interface{}{"type": "text"}       // To search
	fieldTypeDisabled       = map[string]interface{}{"enabled": false}     // No indexing
	fieldTypeTextWithPrefix = map[string]interface{}{"type": "text", "index_prefixes": map[string]interface{}{"min_chars": 1, "max_chars": 10}}
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
			elastic.SetHealthcheck(true),
			elastic.SetBasicAuth(config.Config.ElasticUsername.Get(), config.Config.ElasticPassword.Get()),
		}

		client, err = elastic.NewClient(ops...)
	}

	return client, ctx, err
}

func indexDocument(index string, key string, doc interface{}) error {

	client, ctx, err := GetElastic()
	if err != nil {
		return err
	}

	_, err = client.Index().Index(index).Id(key).BodyJson(doc).Do(ctx)
	return err
}

func indexDocuments(index string, docs map[string]interface{}) error {

	client, ctx, err := GetElastic()
	if err != nil {
		return err
	}

	bulk := client.Bulk()
	for key, doc := range docs {
		bulk.Add(elastic.NewBulkIndexRequest().Index(index).Id(key).Doc(doc))
	}

	resp, err := bulk.Do(ctx)
	if err != nil {
		return err
	}

	failed := resp.Failed()
	if len(failed) > 0 {
		return errors.New(failed[0].Error.Reason)
	}

	return nil
}

func DeleteDocument(index string, key string) error {

	client, ctx, err := GetElastic()
	if err != nil {
		return err
	}

	_, err = client.Delete().Index(index).Id(key).Do(ctx)
	return err
}

func rebuildIndex(index string, mapping map[string]interface{}) {

	client, ctx, err := GetElastic()
	if err != nil {
		log.Info(err)
		return
	}

	log.Info("Deteing " + index)
	resp, err := client.DeleteIndex(index).Do(ctx)
	if err != nil {
		log.Info(err)
		return
	}
	if !resp.Acknowledged {
		log.Info("delete not acknowledged")
		return
	}

	// time.Sleep(time.Second)

	log.Info("Creating " + index)
	createIndexResp, err := client.CreateIndex(index).BodyJson(mapping).Do(ctx)
	if err != nil {
		log.Info(err)
		return
	}

	if !createIndexResp.Acknowledged {
		log.Info("create not acknowledged")
		return
	}
}
