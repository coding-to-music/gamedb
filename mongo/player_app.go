package mongo

import (
	"strconv"

	"github.com/gamedb/website/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlayerApp struct {
	PlayerID     int64              ``
	AppID        int                ``
	AppName      string             ``
	AppIcon      string             ``
	AppTime      int                ``
	AppPrices    map[string]int     ``
	AppPriceHour map[string]float32 ``
}

func (pa PlayerApp) Key() interface{} {
	return strconv.FormatInt(pa.PlayerID, 10) + "-" + strconv.Itoa(pa.AppID)
}

func (pa PlayerApp) BSON() (ret interface{}) {

	var prices = bson.M{}
	for k, v := range pa.AppPrices {
		prices[k] = v
	}

	var pricesHour = bson.M{}
	for k, v := range pa.AppPriceHour {
		pricesHour[k] = v
	}

	return bson.M{
		"_id":             pa.Key(),
		"player_id":       pa.PlayerID,
		"app_id":          pa.AppID,
		"app_name":        pa.AppName,
		"app_icon":        pa.AppIcon,
		"app_time":        pa.AppTime,
		"app_prices":      prices,
		"app_prices_hour": pricesHour,
	}
}

func GetPlayerApps(playerID int64, offset int64, limit bool) (apps []PlayerApp, err error) {

	client, ctx, err := GetMongo()
	if err != nil {
		return apps, err
	}

	o := options.Find().SetSkip(offset).SetSort(bson.M{"app_time": -1})
	if limit {
		o = o.SetLimit(100)
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionPlayerApps)
	cur, err := c.Find(ctx, bson.M{"player_id": playerID}, o)
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
