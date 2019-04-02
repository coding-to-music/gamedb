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
	CreatedAt         time.Time         `bson:"created_at"`
	AppID             int               `bson:"app_id"`
	PackageID         int               `bson:"package_id"`
	Currency          steam.CountryCode `bson:"currency"`
	Name              string            `bson:"name"`
	Icon              string            `bson:"icon"`
	PriceBefore       int               `bson:"price_before"`
	PriceAfter        int               `bson:"price_after"`
	Difference        int               `bson:"difference"`
	DifferencePercent float64           `bson:"difference_percent"`
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

func (price ProductPrice) GetPath() string {

	if price.AppID != 0 {
		return helpers.GetAppPath(price.AppID, price.Name)
	} else if price.PackageID != 0 {
		return helpers.GetPackagePath(price.PackageID, price.Name)
	}

	return ""
}

func (price ProductPrice) GetIcon() string {
	if price.Icon == "" {
		return helpers.DefaultAppIcon
	} else if strings.HasPrefix(price.Icon, "/") || strings.HasPrefix(price.Icon, "http") {
		return price.Icon
	} else {
		return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(price.AppID) + "/" + price.Icon + ".jpg"
	}
}

func (price ProductPrice) GetPercentChange() float64 {

	if math.IsInf(price.DifferencePercent, 0) {
		return 0
	}
	return helpers.RoundFloatTo2DP(price.DifferencePercent)

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

func GetPrices(offset int64, filter interface{}) (prices []ProductPrice, err error) {

	return getProductPrices(filter, offset, 100, false)
}

func getProductPrices(filter interface{}, offset int64, limit int64, sortOrder bool) (prices []ProductPrice, err error) {

	if filter == nil {
		filter = D{}
	}

	client, ctx, err := getMongo()
	if err != nil {
		return prices, err
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionProductPrices.String())

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
