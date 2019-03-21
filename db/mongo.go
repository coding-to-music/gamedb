package db

import (
	"context"

	"github.com/gamedb/website/config"
	"github.com/gamedb/website/log"
	"go.mongodb.org/mongo-driver/bson"
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

	MongoDatabase = config.Config.MongoDatabase
)

type MongoDocument interface {
	ToBSON() interface{}
	GetMongoKey() interface{}
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

func InsertDocument(collection string, document MongoDocument) (resp *mongo.InsertOneResult, err error) {

	// Save to Mongo
	client, ctx, err := GetMongo()
	if err != nil {
		return resp, err
	}

	c := client.Database(MongoDatabase).Collection(collection)
	return c.InsertOne(ctx, document.ToBSON(), &options.InsertOneOptions{})
}

func InsertDocuments(collection string, documents []MongoDocument) (resp *mongo.InsertManyResult, err error) {

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

	f := false
	c := client.Database(MongoDatabase).Collection(collection)
	return c.InsertMany(ctx, many, &options.InsertManyOptions{Ordered: &f})
}

func GetChanges(offset int64) (changes []Change, err error) {

	client, ctx, err := GetMongo()
	if err != nil {
		return changes, err
	}

	c := client.Database(MongoDatabase, &options.DatabaseOptions{}).Collection(CollectionChanges)

	var limit int64 = 100
	cur, err := c.Find(ctx, bson.M{}, &options.FindOptions{
		Limit: &limit,
		Skip:  &offset,
		Sort:  bson.M{"_id": -1},
	})
	if err != nil {
		return changes, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var change Change
		err := cur.Decode(&change)
		log.Err(err)
		changes = append(changes, change)
	}

	return changes, cur.Err()
}
