package mongo

import (
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ProductPrice struct {
	CreatedAt         time.Time
	AppID             int
	PackageID         int
	Currency          steam.CountryCode
	Name              string
	Icon              string
	PriceBefore       int
	PriceAfter        int
	Difference        int
	DifferencePercent float64
}

func (price ProductPrice) Key() interface{} {
	return nil
}

func (price ProductPrice) BSON() (ret interface{}) {

	return bson.M{

	}
}

func GetProductPrices(appID int, packageID int, offset int64) (prices []ProductPrice, err error) {

	client, ctx, err := GetMongo()
	if err != nil {
		return prices, err
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionProductPrices)

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
