package mongo

import (
	"github.com/gamedb/website/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlayerApp struct {
	PlayerID     int64
	AppID        int
	AppName      string
	AppIcon      string
	AppTime      int
	AppPrices    map[string]int
	AppPriceHour map[string]float32
}

func GetPlayerApps(playerID int64, offset int64) (apps []PlayerApp, err error) {

	client, ctx, err := GetMongo()
	if err != nil {
		return apps, err
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionEvents)

	cur, err := c.Find(ctx, bson.M{"player_id": playerID}, options.Find().SetLimit(100).SetSkip(offset).SetSort(bson.M{"app_time": -1}))
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
