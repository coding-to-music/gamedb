package mongo

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"sync"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	MongoDatabase = config.Config.MongoDatabase.Get()

	ErrNoDocuments = mongo.ErrNoDocuments
)

type Document interface {
	BSON() bson.D
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

var (
	mongoClient     *mongo.Client
	mongoCtx        context.Context
	mongoClientLock sync.Mutex
)

func getMongo() (client *mongo.Client, ctx context.Context, err error) {

	mongoClientLock.Lock()
	defer mongoClientLock.Unlock()

	if mongoClient == nil {

		log.Info("Getting Mongo client")

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
func FindOne(collection collection, filter bson.D, sort bson.D, projection bson.M, document Document) (err error) {

	if filter == nil {
		filter = bson.D{}
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

	resp, err = client.Database(MongoDatabase).Collection(collection.String()).InsertOne(ctx, document.BSON(), options.InsertOne())

	return resp, err
}

// Create or update whole document
func ReplaceOne(collection collection, filter bson.D, document Document) (resp *mongo.UpdateResult, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return resp, err
	}

	resp, err = client.Database(MongoDatabase).Collection(collection.String()).ReplaceOne(ctx, filter, document.BSON(), options.Replace().SetUpsert(true))

	return resp, err
}

func DeleteMany(collection collection, filter bson.D) (resp *mongo.DeleteResult, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return resp, nil
	}

	resp, err = client.Database(MongoDatabase, options.Database()).Collection(collection.String()).DeleteMany(ctx, filter)

	return resp, err
}

func DeleteOne(collection collection, filter bson.D) (resp *mongo.DeleteResult, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return resp, err
	}

	resp, err = client.Database(MongoDatabase, options.Database()).Collection(collection.String()).DeleteOne(ctx, filter, options.Delete())

	return resp, err
}

func UpdateManySet(collection collection, filter bson.D, update bson.D) (resp *mongo.UpdateResult, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return resp, nil
	}

	resp, err = client.Database(MongoDatabase, options.Database()).Collection(collection.String()).UpdateMany(ctx, filter, bson.M{"$set": update}, options.Update())

	return resp, err
}

func UpdateManyUnset(collection collection, columns bson.D) (resp *mongo.UpdateResult, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return resp, nil
	}

	resp, err = client.Database(MongoDatabase, options.Database()).Collection(collection.String()).UpdateMany(ctx, bson.M{}, bson.M{"$unset": columns})

	return resp, err
}

func UpdateOne(collection collection, filter bson.D, update bson.D) (resp *mongo.UpdateResult, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return resp, nil
	}

	resp, err = client.Database(MongoDatabase, options.Database()).Collection(collection.String()).UpdateOne(ctx, filter, bson.M{"$set": update}, options.Update())

	return resp, err
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

	resp, err = client.Database(MongoDatabase).Collection(collection.String()).InsertMany(ctx, many, options.InsertMany().SetOrdered(false))

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

func CountDocuments(collection collection, filter bson.D, ttl int32) (count int64, err error) {

	if filter == nil {
		filter = bson.D{}
	}

	item := memcache.MemcacheMongoCount(mongoFilterToMemcacheKey(collection, filter))
	if ttl > 0 {
		item.Expiration = ttl
	}

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return count, err
		}

		if len(filter) == 0 {
			count, err = client.Database(MongoDatabase).Collection(collection.String()).EstimatedDocumentCount(ctx)
		} else {
			count, err = client.Database(MongoDatabase).Collection(collection.String()).CountDocuments(ctx, filter)
		}

		return count, err
	})

	return count, err
}

// Need to close cursor after calling this
func Find(collection collection, offset int64, limit int64, sort bson.D, filter bson.D, projection bson.M, ops *options.FindOptions) (cur *mongo.Cursor, ctx context.Context, err error) {

	if filter == nil {
		filter = bson.D{}
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

func GetRandomRows(collection collection, count int) (cur *mongo.Cursor, ctx context.Context, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return cur, ctx, err
	}

	pipeline := mongo.Pipeline{
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: count}}}},
	}

	c, err := client.Database(MongoDatabase, options.Database()).Collection(collection.String()).Aggregate(ctx, pipeline, options.Aggregate())
	return c, ctx, err
}

func mongoFilterToMemcacheKey(collection collection, filter bson.D) string {

	if filter == nil {
		filter = bson.D{}
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

type index struct {
	V          int            `json:"v"`
	Key        map[string]int `json:"key"`
	Name       string         `json:"name"`
	NS         string         `json:"ns"`
	Background bool           `json:"background"`
}

type count struct {
	ID    int `json:"_id"`
	Count int `json:"count"`
}
