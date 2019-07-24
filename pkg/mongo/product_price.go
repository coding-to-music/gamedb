package mongo

import (
	"errors"
	"math"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ProductPrice struct {
	CreatedAt         time.Time          `bson:"created_at"`
	AppID             int                `bson:"app_id"`
	PackageID         int                `bson:"package_id"`
	Currency          steam.CurrencyCode `bson:"currency"`
	ProdCC            steam.ProductCC    `bson:"prod_cc"`
	Name              string             `bson:"name"`
	Icon              string             `bson:"icon"`
	PriceBefore       int                `bson:"price_before"`
	PriceAfter        int                `bson:"price_after"`
	Difference        int                `bson:"difference"`
	DifferencePercent float64            `bson:"difference_percent"`
}

func (price ProductPrice) BSON() (ret interface{}) {

	return M{
		"created_at":         price.CreatedAt,
		"app_id":             price.AppID,
		"package_id":         price.PackageID,
		"currency":           price.Currency,
		"prod_cc":            price.ProdCC,
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
		return helpers.GetAppPath(price.AppID, price.Name) + "#prices"
	} else if price.PackageID != 0 {
		return helpers.GetPackagePath(price.PackageID, price.Name) + "#prices"
	}

	return ""
}

func (price ProductPrice) GetIcon() string {
	return helpers.GetAppIcon(price.AppID, price.Icon)
}

func (price ProductPrice) GetPercentChange() float64 {

	if math.IsInf(price.DifferencePercent, 0) {
		return 0
	}
	return helpers.RoundFloatTo2DP(price.DifferencePercent)
}

func (price ProductPrice) OutputForJSON() (output []interface{}) {

	return []interface{}{
		price.AppID,                        // 0
		price.PackageID,                    // 1
		price.Currency,                     // 2
		helpers.InsertNewLines(price.Name), // 3
		price.GetIcon(),                    // 4
		price.GetPath(),                    // 5
		helpers.FormatPrice(price.Currency, price.PriceBefore), // 6
		helpers.FormatPrice(price.Currency, price.PriceAfter),  // 7
		helpers.FormatPrice(price.Currency, price.Difference),  // 8
		price.GetPercentChange(),                               // 9
		price.CreatedAt.Format(helpers.DateTime),               // 10
		price.CreatedAt.Unix(),                                 // 11
		price.Difference,                                       // 12 Raw difference
	}
}

func GetPricesByID(IDs []string) (prices []ProductPrice, err error) {

	if IDs == nil || len(IDs) < 1 {
		return prices, nil
	}

	var idsBSON A
	for _, v := range IDs {
		idsBSON = append(idsBSON, v)
	}

	return getProductPrices(M{"_id": M{"$in": idsBSON}}, 0, 0, true)
}

func GetPricesForProduct(productID int, productType helpers.ProductType, cc steam.ProductCC) (prices []ProductPrice, err error) {

	var filter = M{
		"prod_cc": string(cc),
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

func GetPrices(offset int64, limit int64, filter interface{}) (prices []ProductPrice, err error) {

	return getProductPrices(filter, offset, limit, false)
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
		o.SetSort(D{{"created_at", 1}})
	} else {
		o.SetSort(D{{"created_at", -1}})
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
