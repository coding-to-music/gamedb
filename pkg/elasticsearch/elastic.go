package elasticsearch

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/olivere/elastic/v7"
)

var ErrNoResult = errors.New("no result")

const (
	IndexAchievements = "achievements"
	IndexBundles      = "bundles"
	IndexGroups       = "groups"
	IndexArticles     = "articles"
	IndexApps         = "apps"
	IndexGlobal       = "global"
	IndexPlayers      = "players"
)

var (
	settings = map[string]interface{}{
		"number_of_shards":   1,
		"number_of_replicas": 0,
	}

	fieldTypeInt32    = map[string]interface{}{"type": "integer"}
	fieldTypeInt64    = map[string]interface{}{"type": "long"}
	fieldTypeFloat32  = map[string]interface{}{"type": "float"}
	fieldTypeFloat16  = map[string]interface{}{"type": "half_float"}
	fieldTypeBool     = map[string]interface{}{"type": "boolean"}
	fieldTypeKeyword  = map[string]interface{}{"type": "keyword"} // Exact matches, case sensitive
	fieldTypeText     = map[string]interface{}{"type": "text"}    // To search, no sorting, case insensitive
	fieldTypeDisabled = map[string]interface{}{"enabled": false}  // No indexing

	// fieldTypeTextWithPrefix = map[string]interface{}{
	// 	"type": "text",
	// 	"index_prefixes": map[string]interface{}{
	// 		"min_chars": 1,
	// 		"max_chars": 10,
	// 	},
	// }
)

var (
	clientStruct  *elastic.Client
	clientContext context.Context
	clientLock    sync.Mutex
)

func client() (*elastic.Client, context.Context, error) {

	clientLock.Lock()
	defer clientLock.Unlock()

	if clientStruct == nil {

		ops := []elastic.ClientOptionFunc{
			elastic.SetURL(config.C.ElasticAddress),
			elastic.SetSniff(false),
			elastic.SetHealthcheck(true),
			elastic.SetBasicAuth(config.C.ElasticUsername, config.C.ElasticPassword),
		}

		var err error
		clientStruct, err = elastic.NewClient(ops...)
		if err != nil {
			return nil, nil, err
		}

		clientContext = context.Background()
	}

	return clientStruct, clientContext, nil
}

func UpdateDocumentFields(index string, key string, doc map[string]interface{}) error {

	client, ctx, err := client()
	if err != nil {
		return err
	}

	_, err = client.Update().Doc(doc).Index(index).Id(key).Do(ctx)
	return err
}

func indexDocument(index string, key string, doc interface{}) error {

	client, ctx, err := client()
	if err != nil {
		return err
	}

	_, err = client.Index().Index(index).Id(key).BodyJson(doc).Do(ctx)
	return err
}

func indexDocuments(index string, docs map[string]interface{}) error {

	client, ctx, err := client()
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

	client, ctx, err := client()
	if err != nil {
		return err
	}

	_, err = client.Delete().Index(index).Id(key).Do(ctx)
	return err
}

func rebuildIndex(index string, mapping map[string]interface{}) {

	client, ctx, err := client()
	if err != nil {
		log.InfoS(err)
		return
	}

	log.Info("Deteing " + index)
	resp, err := client.DeleteIndex(index).Do(ctx)
	if err != nil {
		log.InfoS(err)
	} else if !resp.Acknowledged {
		log.Info("delete not acknowledged")
		return
	}

	time.Sleep(time.Second)

	log.Info("Creating " + index)
	createIndexResp, err := client.CreateIndex(index).BodyJson(mapping).Do(ctx)
	if err != nil {
		log.InfoS(err)
		return
	}

	if !createIndexResp.Acknowledged {
		log.Info("create not acknowledged")
		return
	}

	log.Info("Indexes rebuilt")
}
