package mongo

import (
	"context"
	"sync"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo/helpers/logging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var ErrNoDocuments = mongo.ErrNoDocuments

type Document interface {
	BSON() bson.D
}

type collection string

func (c collection) String() string {
	return string(c)
}

const (
	CollectionAppAchievements     collection = "app_achievements"
	CollectionAppArticles         collection = "app_articles"
	CollectionAppDLC              collection = "app_dlc"
	CollectionAppItems            collection = "app_items"
	CollectionApps                collection = "apps"
	CollectionAppSales            collection = "app_offers"
	CollectionAppSameOwners       collection = "app_same_owners"
	CollectionBundlePrices        collection = "bundle_prices"
	CollectionChangeItems         collection = "change_products"
	CollectionChanges             collection = "changes"
	CollectionChatBotCommands     collection = "chat_bot_commands"
	CollectionDelayQueue          collection = "delay_queue"
	CollectionDiscordGuilds       collection = "discord_guilds"
	CollectionEvents              collection = "events"
	CollectionGroups              collection = "groups"
	CollectionPackageApps         collection = "package_apps"
	CollectionPackages            collection = "packages"
	CollectionWebhooks            collection = "patreon_webhooks"
	CollectionPlayerAchievements  collection = "player_achievements"
	CollectionPlayerAliases       collection = "player_aliases"
	CollectionPlayerApps          collection = "player_apps"
	CollectionPlayerAppsRecent    collection = "player_apps_recent"
	CollectionPlayerBadges        collection = "player_badges"
	CollectionPlayerBadgesSummary collection = "player_badges_summary"
	CollectionPlayerFriends       collection = "player_friends"
	CollectionPlayerGroups        collection = "player_groups"
	CollectionPlayers             collection = "players"
	CollectionPlayerWishlistApps  collection = "player_wishlist_apps"
	CollectionProductPrices       collection = "product_prices"
	CollectionStats               collection = "stats"
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

		if config.C.TwitchClientID == "" || config.C.TwitchClientSecret == "" {
			return nil, nil, config.ErrMissingEnvironmentVariable
		}

		ctx = context.Background()

		creds := options.Credential{
			AuthSource:  config.C.MongoDatabase,
			Username:    config.C.MongoUsername,
			Password:    config.C.MongoPassword,
			PasswordSet: true,
		}

		ops := options.Client().
			SetAuth(creds).
			ApplyURI(config.MongoDSN()).
			SetAppName("Global Steam")

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

func closeCursor(cur *mongo.Cursor, ctx context.Context) {
	err := cur.Close(ctx)
	if err != nil {
		log.ErrS(err)
	}
}

func EnsureIndexes() {
	log.Info("Starting migrations")
	ensureAppIndexes()
	ensureGroupIndexes()
	ensurePackageIndexes()
	ensurePlayerIndexes()
	ensurePlayerAchievementIndexes()
	ensurePlayerAppIndexes()
	ensurePlayerFriendIndexes()
	ensureSaleIndexes()
	ensureStatIndexes()
	ensureAppSameOwnersIndexes()
	log.Info("Finished migrations")
}

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

	ql := logging.NewLogger("FindOne", collection.String(), filter, sort)

	result := client.Database(config.C.MongoDatabase).
		Collection(collection.String()).
		FindOne(ctx, filter, ops)

	ql.End()

	err = result.Err()
	if err != nil {
		return err
	}

	return result.Decode(document)
}

func GetDistict(collection collection, field string) (values []interface{}, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return values, err
	}

	return client.Database(config.C.MongoDatabase).Collection(collection.String()).Distinct(ctx, field, bson.D{})
}

// Errors if key already exists
func InsertOne(collection collection, document Document) (resp *mongo.InsertOneResult, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return resp, err
	}

	resp, err = client.Database(config.C.MongoDatabase).
		Collection(collection.String()).
		InsertOne(ctx, document.BSON(), options.InsertOne())

	return resp, err
}

// Create or update whole document
func ReplaceOne(collection collection, filter bson.D, document Document) (resp *mongo.UpdateResult, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return resp, err
	}

	resp, err = client.Database(config.C.MongoDatabase).
		Collection(collection.String()).
		ReplaceOne(ctx, filter, document.BSON(), options.Replace().SetUpsert(true))

	return resp, err
}

func DeleteMany(collection collection, filter bson.D) (resp *mongo.DeleteResult, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return resp, nil
	}

	resp, err = client.Database(config.C.MongoDatabase, options.Database()).
		Collection(collection.String()).
		DeleteMany(ctx, filter)

	return resp, err
}

//noinspection GoUnusedExportedFunction
func DeleteOne(collection collection, filter bson.D) (resp *mongo.DeleteResult, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return resp, err
	}

	resp, err = client.Database(config.C.MongoDatabase, options.Database()).
		Collection(collection.String()).
		DeleteOne(ctx, filter, options.Delete())

	return resp, err
}

// Does NOT upsert
func UpdateManySet(collection collection, filter bson.D, update bson.D) (resp *mongo.UpdateResult, err error) {

	if filter == nil {
		filter = bson.D{}
	}

	client, ctx, err := getMongo()
	if err != nil {
		return resp, nil
	}

	resp, err = client.Database(config.C.MongoDatabase, options.Database()).
		Collection(collection.String()).
		UpdateMany(ctx, filter, bson.M{"$set": update}, options.Update())

	return resp, err
}

//noinspection GoUnusedExportedFunction
func UpdateManyUnset(collection collection, columns bson.D) (resp *mongo.UpdateResult, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return resp, nil
	}

	resp, err = client.Database(config.C.MongoDatabase, options.Database()).
		Collection(collection.String()).
		UpdateMany(ctx, bson.M{}, bson.M{"$unset": columns})

	return resp, err
}

// Does not upsert
func UpdateOne(collection collection, filter bson.D, update bson.D) (resp *mongo.UpdateResult, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return resp, nil
	}

	resp, err = client.Database(config.C.MongoDatabase, options.Database()).
		Collection(collection.String()).
		UpdateOne(ctx, filter, bson.M{"$set": update}, options.Update())

	return resp, err
}

// DOES upsert
func UpdateOneWithInsert(collection collection, filter bson.D, update bson.D, onInsert bson.D) (resp *mongo.UpdateResult, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return resp, nil
	}

	resp, err = client.Database(config.C.MongoDatabase, options.Database()).
		Collection(collection.String()).
		UpdateOne(ctx, filter, bson.M{"$set": update, "$setOnInsert": onInsert}, options.Update().SetUpsert(true))

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

	resp, err = client.Database(config.C.MongoDatabase).
		Collection(collection.String()).
		InsertMany(ctx, many, options.InsertMany().SetOrdered(false))

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

func CountDocuments(collection collection, filter bson.D, ttl uint32) (count int64, err error) {

	item := memcache.ItemMongoCount(collection.String(), filter)
	if ttl > 0 {
		item.Expiration = ttl
	}

	err = memcache.GetSetInterface(item, &count, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return count, err
		}

		ql := logging.NewLogger("Count", collection.String(), filter, nil)

		if len(filter) == 0 {
			count, err = client.Database(config.C.MongoDatabase).
				Collection(collection.String()).
				EstimatedDocumentCount(ctx)
		} else {
			count, err = client.Database(config.C.MongoDatabase).
				Collection(collection.String()).
				CountDocuments(ctx, filter)
		}

		ql.End()

		return count, err
	})

	return count, err
}

// Need to close cursor after calling this
func find(collection collection, offset int64, limit int64, sort bson.D, filter bson.D, projection bson.M, ops *options.FindOptions) (cur *mongo.Cursor, ctx context.Context, err error) {

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
	if projection != nil && len(projection) > 0 {
		ops.SetProjection(projection)
	}

	ql := logging.NewLogger("Find", collection.String(), filter, sort)

	cur, err = client.Database(config.C.MongoDatabase, options.Database()).
		Collection(collection.String()).
		Find(ctx, filter, ops)

	ql.End()

	return cur, ctx, err
}

func GetRandomRows(collection collection, count int, filter bson.D, projection bson.M) (cur *mongo.Cursor, ctx context.Context, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return cur, ctx, err
	}

	pipeline := mongo.Pipeline{}

	if len(filter) > 0 {
		pipeline = append(pipeline, bson.D{{"$match", filter}})
	}

	// Must be after filter
	pipeline = append(pipeline, bson.D{{"$sample", bson.D{{Key: "size", Value: count}}}})

	if len(projection) > 0 {
		pipeline = append(pipeline, bson.D{{"$project", projection}})
	}

	//
	ql := logging.NewLogger("Random Aggregate", collection.String(), pipeline, nil)

	c, err := client.Database(config.C.MongoDatabase, options.Database()).
		Collection(collection.String()).
		Aggregate(ctx, pipeline, options.Aggregate())

	ql.End()

	return c, ctx, err
}

type Count struct {
	ID    int `json:"id" bson:"_id"`
	Count int `json:"count"`
}

type StringCount struct {
	ID    string `json:"id" bson:"_id"`
	Count int    `json:"count"`
}

type DateCount struct {
	Date  string `json:"date" bson:"_id"`
	Count int    `json:"count"`
}

func Close() {

	client, ctx, err := getMongo()
	if err != nil {
		log.ErrS(err)
		return
	}

	if client.NumberSessionsInProgress() > 0 {

		err = client.Disconnect(ctx)
		if err != nil {
			log.ErrS(err)
		}
	}
}
