package db

import (
	"context"
	"time"

	"github.com/gamedb/website/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	collectionPlayerApps = "player_apps"
	collectionEvents     = "events"
)

var (
	mongoClient *mongo.Client
	mongoCtx    context.Context

	database = config.Config.MongoDatabase
)

func GetMongo() (client *mongo.Client, ctx context.Context, err error) {

	if mongoClient == nil {

		ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)

		creds := options.Credential{
			AuthSource:  database,
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
