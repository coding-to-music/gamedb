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
	CollectionPlayerApps    = "player_apps"
	CollectionEvents        = "events"
	CollectionChanges       = "changes"
	CollectionProductPrices = "app_pricess"
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

func CountDocuments(collection string, filter interface{}) (count int64, err error) {

	client, ctx, err := GetMongo()
	if err != nil {
		return count, err
	}

	if filter == nil {
		filter = bson.M{}
	}

	c := client.Database(MongoDatabase).Collection(collection)
	return c.CountDocuments(ctx, filter, &options.CountOptions{})
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

func GetEvents(playerID int64, offset int64) (events []Event, err error) {

	client, ctx, err := GetMongo()
	if err != nil {
		return events, err
	}

	c := client.Database(MongoDatabase, &options.DatabaseOptions{}).Collection(CollectionEvents)

	var limit int64 = 100
	cur, err := c.Find(ctx, bson.M{"player_id": playerID}, &options.FindOptions{
		Limit: &limit,
		Skip:  &offset,
		Sort:  bson.M{"created_at": -1},
	})
	if err != nil {
		return events, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var event Event
		err := cur.Decode(&event)
		log.Err(err)
		events = append(events, event)
	}

	return events, cur.Err()
}

func GetNews(appID int, offset int64) (news []News, err error) {

	client, ctx, err := GetMongo()
	if err != nil {
		return news, err
	}

	c := client.Database(MongoDatabase, &options.DatabaseOptions{}).Collection(CollectionEvents)

	var limit int64 = 100
	cur, err := c.Find(ctx, bson.M{"app_id": appID}, &options.FindOptions{
		Limit: &limit,
		Skip:  &offset,
		Sort:  bson.M{"created_at": -1},
	})
	if err != nil {
		return news, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var article News
		err := cur.Decode(&article)
		log.Err(err)
		news = append(news, article)
	}

	return news, cur.Err()
}

func GetProductPricesMongo(appID int, packageID int, offset int64) (prices []ProductPrice, err error) {

	client, ctx, err := GetMongo()
	if err != nil {
		return prices, err
	}

	c := client.Database(MongoDatabase, &options.DatabaseOptions{}).Collection(CollectionProductPrices)

	o := options.FindOptions{}
	o.Skip = &offset

	filter := bson.M{}

	var limit int64 = 100

	if appID != 0 {
		filter["app_id"] = appID
		o.Sort = bson.M{"created_at": -1}
	} else if packageID != 0 {
		filter["package_id"] = packageID
		o.Sort = bson.M{"created_at": -1}
	} else {
		o.Limit = &limit
		o.Sort = bson.M{"created_at": 1}
	}

	cur, err := c.Find(ctx, filter, &o)
	if err != nil {
		return prices, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var price ProductPrice
		err := cur.Decode(&price)
		log.Err(err)
		prices = append(prices, price)
	}

	return prices, cur.Err()
}

func GetPlayerApps(playerID int64, offset int64) (apps []PlayerApp, err error) {

	client, ctx, err := GetMongo()
	if err != nil {
		return apps, err
	}

	c := client.Database(MongoDatabase, &options.DatabaseOptions{}).Collection(CollectionEvents)

	var limit int64 = 100
	cur, err := c.Find(ctx, bson.M{"player_id": playerID}, &options.FindOptions{
		Limit: &limit,
		Skip:  &offset,
		Sort:  bson.M{"app_time": -1},
	})
	if err != nil {
		return apps, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var app PlayerApp
		err := cur.Decode(&app)
		log.Err(err)
		apps = append(apps, app)
	}

	return apps, cur.Err()
}
