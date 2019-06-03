package mongo

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"sync"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	mongoClient *mongo.Client
	mongoCtx    context.Context

	MongoDatabase = config.Config.MongoDatabase.Get()

	ErrNoDocuments = mongo.ErrNoDocuments
)

type Document interface {
	BSON() interface{}
}

type (
	D bson.D
	E bson.E
	M bson.M
	A bson.A
)

type collection string

func (c collection) String() string {
	return string(c)
}

const (
	CollectionAppArticles     collection = "app_articles"
	CollectionChanges         collection = "changes"
	CollectionEvents          collection = "events"
	CollectionGroups          collection = "groups"
	CollectionPatreonWebhooks collection = "patreon_webhooks"
	CollectionPlayerApps      collection = "player_apps"
	CollectionPlayerBadges    collection = "player_badges"
	CollectionPlayers         collection = "players"
	CollectionProductPrices   collection = "product_prices"
	CollectionSessions        collection = "sessions"
	CollectionWishlistApps    collection = "wishlist-apps"
	CollectionWishlistTags    collection = "wishlist-tags"
)

var mongoLock sync.Mutex

func getMongo() (client *mongo.Client, ctx context.Context, err error) {

	mongoLock.Lock()
	defer mongoLock.Unlock()

	if mongoClient == nil {

		ctx = context.Background()

		creds := options.Credential{
			AuthSource:  MongoDatabase,
			Username:    config.Config.MongoUsername.Get(),
			Password:    config.Config.MongoPassword.Get(),
			PasswordSet: true,
		}

		client, err = mongo.NewClient(options.Client().SetAuth(creds).ApplyURI(config.MongoDSN()))

		if err != nil {
			return client, ctx, err
		}

		err = client.Connect(ctx)
		if err != nil {
			return client, ctx, err
		}

		err = client.Ping(ctx, readpref.Primary())
		if err != nil {
			return client, ctx, err
		}

		mongoClient = client
		mongoCtx = ctx
	}

	return mongoClient, mongoCtx, err
}

// Returns ErrNoDocuments on nothing found
func FindDocumentByKey(collection collection, col string, val interface{}, projection M, document Document) (err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	ops := options.FindOne()
	if projection != nil {
		ops.SetProjection(projection)
	}

	c := client.Database(MongoDatabase).Collection(collection.String())
	result := c.FindOne(ctx, M{col: val}, ops)

	return result.Decode(document)
}

func GetFirstDocument(collection collection, filter interface{}, sort interface{}, projection M, document Document) (err error) {

	if filter == nil {
		filter = M{}
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	ops := options.FindOne()

	if projection != nil {
		ops.SetProjection(projection)
	}
	if sort != nil {
		ops.SetSort(sort)
	}

	c := client.Database(MongoDatabase).Collection(collection.String())
	result := c.FindOne(ctx, filter, ops)

	return result.Decode(document)
}

// Errors if key already exists
func InsertDocument(collection collection, document Document) (resp *mongo.InsertOneResult, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return resp, err
	}

	c := client.Database(MongoDatabase).Collection(collection.String())
	return c.InsertOne(ctx, document.BSON(), options.InsertOne())
}

// Create or update whole document
func ReplaceDocument(collection collection, filter interface{}, document Document) (resp *mongo.UpdateResult, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return resp, err
	}

	c := client.Database(MongoDatabase).Collection(collection.String())
	return c.ReplaceOne(ctx, filter, document.BSON(), options.Replace().SetUpsert(true))
}

// Will skip documents that already exist
func InsertDocuments(collection collection, documents []Document) (resp *mongo.InsertManyResult, err error) {

	if len(documents) < 1 {
		return resp, nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return resp, err
	}

	var many []interface{}
	for _, v := range documents {
		many = append(many, v.BSON())
	}

	c := client.Database(MongoDatabase).Collection(collection.String())
	resp, err = c.InsertMany(ctx, many, options.InsertMany().SetOrdered(false))

	bulkErr, ok := err.(mongo.BulkWriteException)
	if ok {
		for _, v := range bulkErr.WriteErrors {
			if v.Code != 11000 { // duplicate key
				return resp, err
			}
		}
		return resp, nil
	}

	return resp, err
}

func CountDocuments(collection collection, filter interface{}) (count int64, err error) {

	if filter == nil {
		filter = M{}
	}

	b, err := json.Marshal(filter)
	log.Err(err)

	h := md5.Sum(b)

	key := hex.EncodeToString(h[:])

	item := helpers.MemcacheMongoCount(collection.String() + "-" + key)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return count, err
		}

		c := client.Database(MongoDatabase).Collection(collection.String())

		return c.CountDocuments(ctx, filter, options.Count())
	})

	return count, err
}

func DeleteColumn(collection collection, column string) (err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return nil
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(collection.String())
	_, err = c.UpdateMany(ctx, M{}, M{"$unset": M{column: ""}})
	return err
}

func DeleteRows(collection collection, filter M) (err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return nil
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(collection.String())
	_, err = c.DeleteMany(ctx, filter)
	return err
}
