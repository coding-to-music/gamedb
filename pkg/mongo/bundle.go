package mongo

import (
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"go.mongodb.org/mongo-driver/bson"
)

type Bundle struct {
	Apps            int                        `bson:"apps"`
	Discount        int                        `bson:"discount"`
	DiscountHighest int                        `bson:"discount_highest"`
	DiscountSale    int                        `bson:"discount_sale"`
	Icon            string                     `bson:"icon"`
	ID              int                        `bson:"_id"`
	Name            string                     `bson:"name"`
	NameMarked      string                     `bson:"-"`
	Packages        int                        `bson:"packages"`
	Prices          map[steamapi.ProductCC]int `bson:"prices"`
	PricesSale      map[steamapi.ProductCC]int `bson:"prices_sale"`
	Score           float64                    `bson:"-"`
	Type            string                     `bson:"type"`
	UpdatedAt       time.Time                  `bson:"updated_at"`
}

func (bundle *Bundle) BSON() bson.D {

	bundle.UpdatedAt = time.Now()

	return bson.D{
		{"_id", bundle.ID},
		{"apps", bundle.Apps},
		{"discount", bundle.Discount},
		{"discount_highest", bundle.DiscountHighest},
		{"discount_sale", bundle.DiscountSale},
		{"icon", bundle.Icon},
		{"name", bundle.Name},
		{"packages", bundle.Packages},
		{"prices", bundle.Prices},
		{"prices_sale", bundle.PricesSale},
		{"type", bundle.Type},
		{"updated_at", bundle.UpdatedAt},
	}
}

func BatchBundles(filter bson.D, projection bson.M, callback func(bundles []Bundle)) (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		bundles, err := GetBundles(offset, limit, bson.D{{"_id", 1}}, filter, projection)
		if err != nil {
			return err
		}

		callback(bundles)

		if int64(len(bundles)) != limit {
			break
		}

		offset += limit
	}

	return nil
}

func GetBundle(id int) (bundle Bundle, err error) {

	err = memcache.GetSetInterface(memcache.ItemBundle(id), &bundle, func() (interface{}, error) {

		err = FindOne(CollectionBundles, bson.D{{"_id", id}}, nil, nil, &bundle)
		return bundle, err
	})

	return bundle, err
}

func GetBundles(offset int64, limit int64, sort bson.D, filter bson.D, projection bson.M) (bundles []Bundle, err error) {

	cur, ctx, err := find(CollectionBundles, offset, limit, sort, filter, projection, nil)
	if err != nil {
		return bundles, err
	}

	defer closeCursor(cur, ctx)

	for cur.Next(ctx) {

		var bundle Bundle
		err := cur.Decode(&bundle)
		if err != nil {
			log.ErrS(err, bundle.ID)
		} else {
			bundles = append(bundles, bundle)
		}
	}

	return bundles, cur.Err()
}
