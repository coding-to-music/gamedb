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
	. "go.mongodb.org/mongo-driver/bson"
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

type collection string

func (c collection) String() string {
	return string(c)
}

const (
	CollectionAppArticles         collection = "app_articles"
	CollectionAppItems            collection = "app_items"
	CollectionAppSales            collection = "app_offers"
	CollectionApps                collection = "apps"
	CollectionBundlePrices        collection = "bundle_prices"
	CollectionChangeItems         collection = "change_products"
	CollectionChanges             collection = "changes"
	CollectionEvents              collection = "events"
	CollectionGroups              collection = "groups"
	CollectionPatreonWebhooks     collection = "patreon_webhooks"
	CollectionPackageApps         collection = "package_apps"
	CollectionPlayerApps          collection = "player_apps"
	CollectionPlayerAppsRecent    collection = "player_apps_recent"
	CollectionPlayerBadges        collection = "player_badges"
	CollectionPlayerBadgesSummary collection = "player_badges_summary"
	CollectionPlayerFriends       collection = "player_friends"
	CollectionPlayerGroups        collection = "player_groups"
	CollectionPlayerWishlistApps  collection = "player_wishlist_apps"
	CollectionPlayers             collection = "players"
	CollectionProductPrices       collection = "product_prices"
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
func FindOne(collection collection, filter D, sort D, projection M, document Document) (err error) {

	if filter == nil {
		filter = D{}
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

	ql := helpers.QueryLogger{}
	ql.Start("FindOne", collection.String(), filter, sort)

	result := client.Database(MongoDatabase).Collection(collection.String()).FindOne(ctx, filter, ops)

	ql.End()

	return result.Decode(document)
}

// Errors if key already exists
func InsertOne(collection collection, document Document) (resp *mongo.InsertOneResult, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return resp, err
	}

	ql := helpers.QueryLogger{}
	ql.Start("InsertOne", collection.String(), nil, nil)

	r, err := client.Database(MongoDatabase).Collection(collection.String()).InsertOne(ctx, document.BSON(), options.InsertOne())

	ql.End()

	return r, err
}

// Create or update whole document
func ReplaceOne(collection collection, filter D, document Document) (resp *mongo.UpdateResult, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return resp, err
	}

	ql := helpers.QueryLogger{}
	ql.Start("ReplaceOne", collection.String(), filter, nil)

	r, err := client.Database(MongoDatabase).Collection(collection.String()).ReplaceOne(ctx, filter, document.BSON(), options.Replace().SetUpsert(true))

	ql.End()

	return r, err
}

func DeleteMany(collection collection, filter D) (err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return nil
	}

	ql := helpers.QueryLogger{}
	ql.Start("DeleteMany", collection.String(), filter, nil)

	_, err = client.Database(MongoDatabase, options.Database()).Collection(collection.String()).DeleteMany(ctx, filter)

	ql.End()

	return err
}

func DeleteOne(collection collection, filter D) (err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	ql := helpers.QueryLogger{}
	ql.Start("DeleteOne", collection.String(), filter, nil)

	_, err = client.Database(MongoDatabase, options.Database()).Collection(collection.String()).DeleteOne(ctx, filter, options.Delete())

	ql.End()

	return err
}

func UpdateManySet(collection collection, filter D, update D) (result *mongo.UpdateResult, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return result, nil
	}

	ql := helpers.QueryLogger{}
	ql.Start("UpdateMany", collection.String(), filter, nil)

	_, err = client.Database(MongoDatabase, options.Database()).Collection(collection.String()).UpdateMany(ctx, filter, M{"$set": update}, options.Update())

	ql.End()

	return result, err
}

func UpdateManyUnset(collection collection, columns D) (err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return nil
	}

	ql := helpers.QueryLogger{}
	ql.Start("UpdateMany", collection.String(), nil, nil)

	_, err = client.Database(MongoDatabase, options.Database()).Collection(collection.String()).UpdateMany(ctx, M{}, M{"$unset": columns})

	ql.End()

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

	ql := helpers.QueryLogger{}
	ql.Start("InsertMany", collection.String(), nil, nil)

	resp, err = client.Database(MongoDatabase).Collection(collection.String()).InsertMany(ctx, many, options.InsertMany().SetOrdered(false))

	ql.End()

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

func CountDocuments(collection collection, filter D, ttl int32) (count int64, err error) {

	if filter == nil {
		filter = D{}
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
		ql.Start("CountDocuments", collection.String(), filter, nil)

		count, err = client.Database(MongoDatabase).Collection(collection.String()).CountDocuments(ctx, filter, options.Count())

		ql.End()

		return count, err
	})

	return count, err
}

// Need to close cursor after calling this
func Find(collection collection, offset int64, limit int64, sort D, filter D, projection M, ops *options.FindOptions) (cur *mongo.Cursor, ctx context.Context, err error) {

	if filter == nil {
		filter = D{}
	}

	client, ctx, err := getMongo()
	if err != nil {
		return cur, ctx, err
	}

	if ops == nil {
		ops = options.Find()
	}
	if offset > 0 {
		ops.SetSkip(offset)
	}
	if limit > 0 {
		ops.SetLimit(limit)
	}
	if sort != nil {
		ops.SetSort(sort)
	}
	if projection != nil {
		ops.SetProjection(projection)
	}

	ql := helpers.QueryLogger{}
	ql.Start("Find", collection.String(), filter, sort)

	cur, err = client.Database(MongoDatabase, options.Database()).Collection(collection.String()).Find(ctx, filter, ops)

	ql.End()

	return cur, ctx, err
}

func mongoFilterToMemcacheKey(collection collection, filter D) string {

	if filter == nil {
		filter = D{}
	}

	b, err := json.Marshal(filter)
	log.Err(err)

	h := md5.Sum(b)

	key := hex.EncodeToString(h[:])

	return collection.String() + "-" + key
}

func ChunkWriteModels(models []mongo.WriteModel, size int) (chunks [][]mongo.WriteModel) {

	for i := 0; i < len(models); i += size {
		end := i + size

		if end > len(models) {
			end = len(models)
		}

		chunks = append(chunks, models[i:end])
	}
	return chunks
}
