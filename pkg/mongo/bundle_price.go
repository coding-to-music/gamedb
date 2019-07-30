package mongo

import (
	"strconv"
	"time"
)

type BundlePrice struct {
	CreatedAt time.Time `bson:"created_at"`
	BundleID  int       `bson:"bundle_id"`
	Discount  int       `bson:"bundles"`
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
