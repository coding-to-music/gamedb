package mongo

import (
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BundlePrice struct {
	CreatedAt time.Time `bson:"created_at"`
	BundleID  int       `bson:"bundle_id"`
	Discount  int       `bson:"price"`
}

func (price BundlePrice) BSON() (ret interface{}) {

	return M{
		"_id":        price.getKey(),
		"created_at": price.CreatedAt,
		"bundle_id":  price.BundleID,
		"price":      price.Discount,
	}
}

func (price BundlePrice) getKey() string {
	return strconv.Itoa(price.BundleID) + "-" + price.CreatedAt.Format(time.RFC3339)
}

func GetBundlePrices(bundleID int) (prices []BundlePrice, err error) {

	filter := M{"bundle_id": bundleID}

	client, ctx, err := getMongo()
	if err != nil {
		return prices, err
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionBundlePrices.String())

	ops := options.Find()
	ops.SetSort(D{{"created_at", 1}})

	cur, err := c.Find(ctx, filter, ops)
	if err != nil {
		return prices, err
	}

	defer func(cur *mongo.Cursor) {
		err = cur.Close(ctx)
		log.Err(err)
	}(cur)

	for cur.Next(ctx) {

		var price BundlePrice
		err := cur.Decode(&price)
		if err != nil {
			log.Err(err, price.getKey())
		}
		prices = append(prices, price)
	}

	return prices, cur.Err()
}
