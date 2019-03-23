package mongo

import (
	"context"

	"github.com/gamedb/website/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	CollectionChanges       = "changes"
	CollectionEvents        = "events"
	CollectionPlayers       = "players"
	CollectionPlayerApps    = "player_apps"
	CollectionProductPrices = "app_pricess"
)

var (
	mongoClient *mongo.Client
	mongoCtx    context.Context

	MongoDatabase = config.Config.MongoDatabase
)

type MongoDocument interface {
	BSON() interface{}
	Key() interface{}
}

func GetMongo() (client *mongo.Client, ctx context.Context, err error) {

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

// Errors if key already exists
func InsertDocument(collection string, document MongoDocument) (resp *mongo.InsertOneResult, err error) {

	// Save to Mongo
	client, ctx, err := GetMongo()
	if err != nil {
		return resp, err
	}

	c := client.Database(MongoDatabase).Collection(collection)
	return c.InsertOne(ctx, document.BSON(), options.InsertOne())
}

// Create or update whole document
func ReplaceDocument(collection string, filter interface{}, document MongoDocument) (resp *mongo.UpdateResult, err error) {

	client, ctx, err := GetMongo()
	if err != nil {
		return resp, err
	}

	c := client.Database(MongoDatabase).Collection(collection)
	return c.ReplaceOne(ctx, filter, document.BSON(), options.Replace().SetUpsert(true))
}

// Will skip documents that already exist
func InsertDocuments(collection string, documents []MongoDocument) (resp *mongo.InsertManyResult, err error) {

	if len(documents) < 1 {
		return resp, nil
	}

	client, ctx, err := GetMongo()
	if err != nil {
		return resp, err
	}

	var many []interface{}
	for _, v := range documents {
		many = append(many, v.BSON())
	}

	c := client.Database(MongoDatabase).Collection(collection)
	resp, err = c.InsertMany(ctx, many, options.InsertMany().SetOrdered(false))

	serr, ok := err.(mongo.BulkWriteException)
	if ok {
		for _, v := range serr.WriteErrors {
			if v.Code != 11000 { // duplicate key
				return resp, err
			}
		}
		return resp, nil
	}

	return resp, err
}

func CountDocuments(collection string, filter interface{}) (count int64, err error) {

	client, ctx, err := GetMongo()
	if err != nil {
		return count, err
	}

	c := client.Database(MongoDatabase).Collection(collection)
	return c.CountDocuments(ctx, filter, options.Count())
}
