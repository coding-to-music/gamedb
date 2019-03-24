package mongo

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

	ErrNoDocuments = mongo.ErrNoDocuments
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

func FindDocument(collection string, col string, val interface{}, document MongoDocument) (err error) {

	client, ctx, err := GetMongo()
	if err != nil {
		return err
	}

	c := client.Database(MongoDatabase).Collection(collection)
	result := c.FindOne(ctx, bson.M{col: val}, options.FindOne())

	return result.Decode(document)
}

// Errors if key already exists
func InsertDocument(collection string, document MongoDocument) (resp *mongo.InsertOneResult, err error) {

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

func CountDocuments(collection string, filter interface{}) (count int64, err error) {

	client, ctx, err := GetMongo()
	if err != nil {
		return count, err
	}

	c := client.Database(MongoDatabase).Collection(collection)
	return c.CountDocuments(ctx, filter, options.Count())
}

func documentsToArticles(a []MongoDocument) (b []Article) {

	for _, v := range a {

		original, ok := v.(Article)
		if ok {
			b = append(b, original)
		} else {
			log.Info("kind not a struct")
		}
	}

	return b
}

func documentsToPlayers(a []MongoDocument) (b []Player) {

	for _, v := range a {

		original, ok := v.(Player)
		if ok {
			b = append(b, original)
		} else {
			log.Info("kind not a struct")
		}
	}

	return b
}

func documentsToPlayerApps(a []MongoDocument) (b []PlayerApp) {

	for _, v := range a {

		original, ok := v.(PlayerApp)
		if ok {
			b = append(b, original)
		}
	}

	return b
}

func documentsToChanges(a []MongoDocument) (b []Change) {

	for _, v := range a {

		original, ok := v.(Change)
		if ok {
			b = append(b, original)
		}
	}

	return b
}

func documentsToProductPrices(a []MongoDocument) (b []ProductPrice) {

	for _, v := range a {

		original, ok := v.(ProductPrice)
		if ok {
			b = append(b, original)
		}
	}

	return b
}
