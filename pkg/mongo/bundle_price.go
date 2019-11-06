package mongo

import (
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	. "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type BundlePrice struct {
	CreatedAt time.Time `bson:"created_at"`
	BundleID  int       `bson:"bundle_id"`
	Discount  int       `bson:"price"`
}

func (price BundlePrice) BSON() D {

	return D{
		{"_id", price.getKey()},
		{"created_at", price.CreatedAt},
		{"bundle_id", price.BundleID},
		{"price", price.Discount},
	}
}

func (price BundlePrice) getKey() string {
	return strconv.Itoa(price.BundleID) + "-" + price.CreatedAt.Format(time.RFC3339)
}

func GetBundlePrices(bundleID int) (prices []BundlePrice, err error) {

	var sort = D{{"created_at", 1}}
	var filter = D{{"bundle_id", bundleID}}

	cur, ctx, err := Find(CollectionBundlePrices, 0, 0, sort, filter, nil, nil)
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
