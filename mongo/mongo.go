package mongo

import (
	"context"

	"github.com/gamedb/website/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	mongoClient *mongo.Client
	mongoCtx    context.Context

	MongoDatabase = config.Config.MongoDatabase

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
	CollectionAppArticles   collection = "app_articles"
	CollectionChanges       collection = "changes"
	CollectionEvents        collection = "events"
	CollectionPlayers       collection = "players"
	CollectionPlayerApps    collection = "player_apps"
	CollectionProductPrices collection = "product_prices"
)

func getMongo() (client *mongo.Client, ctx context.Context, err error) {

	if mongoClient == nil {

		ctx = context.Background()

		creds := options.Credential{
			AuthSource:  MongoDatabase,
			Username:    config.Config.MongoUsername,
			Password:    config.Config.MongoPassword,
			PasswordSet: true,
		}

		client, err = mongo.NewClient(options.Client().SetAuth(creds).ApplyURI(config.Config.MongoDSN()))

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

func FindDocument(collection collection, col string, val interface{}, document Document) (err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	c := client.Database(MongoDatabase).Collection(collection.String())
	result := c.FindOne(ctx, bson.M{col: val}, options.FindOne())

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

	// todo, build in cache by hashing the filter

	client, ctx, err := getMongo()
	if err != nil {
		return count, err
	}

	if filter == nil {
		filter = bson.M{}
	}

	c := client.Database(MongoDatabase).Collection(collection.String())
	return c.CountDocuments(ctx, filter, options.Count())
}
