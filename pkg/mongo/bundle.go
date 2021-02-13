package mongo

import (
	"math"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"go.mongodb.org/mongo-driver/bson"
)

const BundleTypeCompleteTheSet = "cts"
const BundleTypePurchaseTogether = "pt"

type Bundle struct {
	Apps            []int                      `bson:"apps"`
	CreatedAt       time.Time                  `bson:"created_at"`
	Discount        int                        `bson:"discount"`
	DiscountHighest int                        `bson:"discount_highest"`
	DiscountLowest  int                        `bson:"discount_lowest"`
	DiscountSale    int                        `bson:"discount_sale"`
	Giftable        bool                       `bson:"giftable"`
	Icon            string                     `bson:"icon"`
	ID              int                        `bson:"_id"`
	Image           string                     `bson:"image"`
	Name            string                     `bson:"name"`
	OnSale          bool                       `bson:"on_sale"`
	Packages        []int                      `bson:"packages"`
	Prices          map[steamapi.ProductCC]int `bson:"prices"`
	PricesSale      map[steamapi.ProductCC]int `bson:"prices_sale"`
	Type            string                     `bson:"type"`
	UpdatedAt       time.Time                  `bson:"updated_at"`
}

func (bundle Bundle) BSON() bson.D {

	bundle.Discount = int(math.Abs(float64(bundle.Discount)))
	bundle.DiscountHighest = int(math.Abs(float64(bundle.DiscountHighest)))
	bundle.DiscountLowest = int(math.Abs(float64(bundle.DiscountLowest)))
	bundle.DiscountSale = int(math.Abs(float64(bundle.DiscountSale)))

	bundle.UpdatedAt = time.Now()

	if bundle.CreatedAt.IsZero() || bundle.CreatedAt.Unix() == 0 {
		bundle.CreatedAt = time.Now()
	}

	bundle.OnSale = bundle.DiscountSale > bundle.Discount

	if bundle.DiscountSale < bundle.Discount {
		bundle.DiscountSale = bundle.Discount
	}

	if bundle.DiscountSale > bundle.DiscountHighest {
		bundle.DiscountHighest = bundle.DiscountSale
	}

	if bundle.Discount < bundle.DiscountLowest {
		bundle.DiscountLowest = bundle.Discount
	}

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
