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
	mongoClient     *mongo.Client
	mongoClientLock sync.Mutex
	mongoCtx        context.Context

	MongoDatabase = config.Config.MongoDatabase.Get()

	ErrNoDocuments = mongo.ErrNoDocuments
)

type Document interface {
	BSON() interface{}
}

type (
	D bson.D
	M bson.M
	A bson.A
	// E bson.E
)

type collection string

func (c collection) String() string {
	return string(c)
}

const (
	CollectionAppArticles        collection = "app_articles"
	CollectionAppItems           collection = "app_items"
	CollectionAppOffers          collection = "app_offers"
	CollectionApps               collection = "apps"
	CollectionBundlePrices       collection = "bundle_prices"
	CollectionChangeItems        collection = "change_products"
	CollectionChanges            collection = "changes"
	CollectionEvents             collection = "events"
	CollectionGroups             collection = "groups"
	CollectionPatreonWebhooks    collection = "patreon_webhooks"
	CollectionPackageApps        collection = "package_apps"
	CollectionPlayerApps         collection = "player_apps"
	CollectionPlayerAppsRecent   collection = "player_apps_recent"
	CollectionPlayerBadges       collection = "player_badges"
	CollectionPlayerFriends      collection = "player_friends"
	CollectionPlayerGroups       collection = "player_groups"
	CollectionPlayerWishlistApps collection = "player_wishlist_apps"
	CollectionPlayers            collection = "players"
	CollectionProductPrices      collection = "product_prices"
	// CollectionWishlistApps       collection = "wishlist-apps"
	// CollectionWishlistTags       collection = "wishlist-tags"

	// Stats
	CollectionTags       collection = "tags"
	CollectionGenres     collection = "genres"
	CollectionDevelopers collection = "developers"
	CollectionPublishers collection = "publishers"
	CollectionCategories collection = "categories"
)

func getMongo() (client *mongo.Client, ctx context.Context, err error) {

	mongoClientLock.Lock()
	defer mongoClientLock.Unlock()

	if mongoClient == nil {

		ctx = context.Background()

		creds := options.Credential{
			AuthSource:  MongoDatabase,
			Username:    config.Config.MongoUsername.Get(),
			Password:    config.Config.MongoPassword.Get(),
			PasswordSet: true,
		}

		ops := options.Client().
			SetAuth(creds).
			ApplyURI(config.MongoDSN()).
			SetAppName("Game DB")

		client, err = mongo.NewClient(ops)

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

// Does not return error on nothing returned
func FindOne(collection collection, filter interface{}, sort interface{}, projection M, document Document) (err error) {

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
func InsertOne(collection collection, document Document) (resp *mongo.InsertOneResult, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return resp, err
	}

	c := client.Database(MongoDatabase).Collection(collection.String())
	return c.InsertOne(ctx, document.BSON(), options.InsertOne())
}

// Create or update whole document
func ReplaceOne(collection collection, filter interface{}, document Document) (resp *mongo.UpdateResult, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return resp, err
	}

	c := client.Database(MongoDatabase).Collection(collection.String())
	return c.ReplaceOne(ctx, filter, document.BSON(), options.Replace().SetUpsert(true))
}

func DeleteMany(collection collection, filter M) (err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return nil
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(collection.String())
	_, err = c.DeleteMany(ctx, filter)
	return err
}

func DeleteOne(collection collection, filter M) (err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(collection.String())

	_, err = c.DeleteOne(ctx, filter, options.Delete())
	return err
}

func UpdateManySet(collection collection, filter M, update M) (result *mongo.UpdateResult, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return result, nil
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(collection.String())
	_, err = c.UpdateMany(ctx, filter, M{"$set": update}, options.Update())
	return result, err
}

// M{column_to_delete: 1}
func UpdateManyUnset(collection collection, columns M) (err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return nil
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(collection.String())
	_, err = c.UpdateMany(ctx, M{}, M{"$unset": columns})
	return err
}

// Will skip documents that already exist
func InsertMany(collection collection, documents []Document) (resp *mongo.InsertManyResult, err error) {

	if documents == nil || len(documents) < 1 {
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

func CountDocuments(collection collection, filter interface{}, ttl int32) (count int64, err error) {

	if filter == nil {
		filter = M{}
	}

	item := helpers.MemcacheMongoCount(mongoFilterToMemcacheKey(collection, filter))
	if ttl > 0 {
		item.Expiration = ttl
	}

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return count, err
		}

		ql := helpers.QueryLogger{}
		ql.Start(collection.String(), filter, nil)

		c, err := client.Database(MongoDatabase).Collection(collection.String()).CountDocuments(ctx, filter, options.Count())

		ql.End()

		return c, err
	})

	return count, err
}

func SetCountDocuments(collection collection, filter interface{}, ttl int32) error {

	item := helpers.MemcacheMongoCount(mongoFilterToMemcacheKey(collection, filter))
	if ttl > 0 {
		item.Expiration = ttl
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	count, err := client.Database(MongoDatabase).Collection(collection.String()).CountDocuments(ctx, filter, options.Count())
	if err != nil {
		return err
	}

	return helpers.GetMemcache().SetInterface(item.Key, count, ttl)
}

func mongoFilterToMemcacheKey(collection collection, filter interface{}) string {

	if filter == nil {
		filter = M{}
	}

	b, err := json.Marshal(filter)
	log.Err(err)

	h := md5.Sum(b)

	key := hex.EncodeToString(h[:])

	return collection.String() + "-" + key
}
