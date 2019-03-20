package db

import (
	"context"

	"github.com/gamedb/website/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	CollectionPlayerApps = "player_apps"
	CollectionEvents     = "events"
	CollectionChanges    = "changes"
)

var (
	mongoClient *mongo.Client
	mongoCtx    context.Context

	mongoDatabase = config.Config.MongoDatabase
)

type MongoDocument interface {
	ToBSON() interface{}
}

func GetMongo() (client *mongo.Client, ctx context.Context, err error) {

	if mongoClient == nil {

		ctx = context.Background()

		creds := options.Credential{
			AuthSource:  mongoDatabase,
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

func InsertDocument(document MongoDocument, collection string) (resp *mongo.InsertOneResult, err error) {

	// Save to Mongo
	client, ctx, err := GetMongo()
	if err != nil {
		return resp, err
	}

	c := client.Database(mongoDatabase).Collection(collection)
	return c.InsertOne(ctx, document.ToBSON())
}

func InsertDocuments(documents []MongoDocument, collection string) (resp *mongo.InsertManyResult, err error) {

	if len(documents) < 1 {
		return resp, nil
	}

	// Save to Mongo
	client, ctx, err := GetMongo()
	if err != nil {
		return resp, err
	}

	var many []interface{}
	for _, v := range documents {
		many = append(many, v.ToBSON())
	}

	c := client.Database(mongoDatabase).Collection(collection)
	return c.InsertMany(ctx, many)
}
