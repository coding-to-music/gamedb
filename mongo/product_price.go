package mongo

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/helpers"
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
		"created_at":         price.CreatedAt,
		"app_id":             price.AppID,
		"package_id":         price.PackageID,
		"currency":           price.Currency,
		"name":               price.Name,
		"icon":               price.Icon,
		"price_before":       price.PriceBefore,
		"price_after":        price.PriceAfter,
		"difference":         price.Difference,
		"difference_percent": price.DifferencePercent,
	}
}

func (p ProductPrice) GetPath() string {

	if p.AppID != 0 {
		return helpers.GetAppPath(p.AppID, p.Name)
	} else if p.PackageID != 0 {
		return helpers.GetPackagePath(p.PackageID, p.Name)
	}

	return ""
}

func (p ProductPrice) GetIcon() string {
	if p.Icon == "" {
		return helpers.DefaultAppIcon
	} else if strings.HasPrefix(p.Icon, "/") || strings.HasPrefix(p.Icon, "http") {
		return p.Icon
	} else {
		return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(p.AppID) + "/" + p.Icon + ".jpg"
	}
}

func (p ProductPrice) GetPercentChange() float64 {

	if math.IsInf(p.DifferencePercent, 0) {
		return 0
	}
	return helpers.RoundFloatTo2DP(p.DifferencePercent)

}

func (price ProductPrice) OutputForJSON() (output []interface{}) {

	locale, err := helpers.GetLocaleFromCountry(price.Currency)
	log.Err(err)

	return []interface{}{
		price.AppID,                              // 0
		price.PackageID,                          // 1
		price.Currency,                           // 2
		price.Name,                               // 3
		price.GetIcon(),                          // 4
		price.GetPath(),                          // 5
		locale.Format(price.PriceBefore),         // 6
		locale.Format(price.PriceAfter),          // 7
		locale.Format(price.Difference),          // 8
		price.GetPercentChange(),                 // 9
		price.CreatedAt.Format(helpers.DateTime), // 10
		price.CreatedAt.Unix(),                   // 11
		price.Difference,                         // 12 Raw difference
	}
}

func CountPrices() (count int64, err error) {

	var item = helpers.MemcachePricesCount

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		return CountDocuments(CollectionProductPrices, bson.M{})
	})

	return count, err
}

func GetPricesForProduct(productID int, productType helpers.ProductType, cc steam.CountryCode) (prices []ProductPrice, err error) {

	var filter = bson.M{
		"currency": string(cc),
	}

	if productType == helpers.ProductTypeApp {
		filter["app_id"] = productID
	} else if productType == helpers.ProductTypePackage {
		filter["package_id"] = productID
	} else {
		return prices, errors.New("invalid product type")
	}

	return getProductPrices(filter, 0, 0, true)
}

func GetPrices(offset int64, cc steam.CountryCode) (prices []ProductPrice, err error) {

	var filter = bson.M{
		"currency": string(cc),
	}

	return getProductPrices(filter, offset, 100, false)
}

func getProductPrices(filter interface{}, offset int64, limit int64, sortOrder bool) (prices []ProductPrice, err error) {

	client, ctx, err := GetMongo()
	if err != nil {
		return prices, err
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionProductPrices)

	o := options.Find()
	o.SetSkip(offset)
	if limit > 0 {
		o.SetLimit(limit)
	}

	if sortOrder {
		o.SetSort(bson.M{"created_at": 1})
	} else {
		o.SetSort(bson.M{"created_at": -1})
	}

	cur, err := c.Find(ctx, filter, o)
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
