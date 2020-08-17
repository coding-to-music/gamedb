package mongo

import (
	"errors"
	"math"
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type ProductPrice struct {
	CreatedAt         time.Time             `bson:"created_at"`
	AppID             int                   `bson:"app_id"`
	PackageID         int                   `bson:"package_id"`
	Currency          steamapi.CurrencyCode `bson:"currency"`
	ProdCC            steamapi.ProductCC    `bson:"prod_cc"`
	Name              string                `bson:"name"`
	Icon              string                `bson:"icon"`
	PriceBefore       int                   `bson:"price_before"`
	PriceAfter        int                   `bson:"price_after"`
	Difference        int                   `bson:"difference"`
	DifferencePercent float64               `bson:"difference_percent"`
}

func (price ProductPrice) BSON() bson.D {

	return bson.D{
		{"created_at", price.CreatedAt},
		{"app_id", price.AppID},
		{"package_id", price.PackageID},
		{"currency", price.Currency},
		{"prod_cc", price.ProdCC},
		{"name", price.Name},
		{"icon", price.Icon},
		{"price_before", price.PriceBefore},
		{"price_after", price.PriceAfter},
		{"difference", price.Difference},
		{"difference_percent", price.DifferencePercent},
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
	icon := helpers.GetAppIcon(price.AppID, price.Icon)
	return strings.TrimPrefix(icon, "https://gamedb.online")
}

func (price ProductPrice) GetPercentChange() float64 {

	if math.IsInf(price.DifferencePercent, 0) {
		return 0 // This is because JSON can not handle infinite
	}
	return helpers.RoundFloatTo2DP(price.DifferencePercent)
}

func (price ProductPrice) OutputForJSON() (output []interface{}) {

	if math.IsInf(price.DifferencePercent, 0) {
		price.DifferencePercent = 0
	}

	return []interface{}{
		price.AppID,     // 0
		price.PackageID, // 1
		price.Currency,  // 2
		price.Name,      // 3
		price.GetIcon(), // 4
		price.GetPath(), // 5
		i18n.FormatPrice(price.Currency, price.PriceBefore), // 6
		i18n.FormatPrice(price.Currency, price.PriceAfter),  // 7
		i18n.FormatPrice(price.Currency, price.Difference),  // 8
		price.GetPercentChange(),                            // 9
		price.CreatedAt.Format(helpers.DateTime),            // 10
		price.CreatedAt.Unix(),                              // 11
		price.Difference,                                    // 12 Raw difference
		price.ProdCC,                                        // 13
		price.DifferencePercent,                             // 14
		math.Round(price.DifferencePercent),                 // 15
	}
}

func GetPricesByID(IDs []string) (prices []ProductPrice, err error) {

	if IDs == nil || len(IDs) < 1 {
		return prices, nil
	}

	var idsBSON bson.A
	for _, ID := range IDs {

		objectID, err := primitive.ObjectIDFromHex(ID)
		zap.S().Error(err)
		if err == nil {
			idsBSON = append(idsBSON, objectID)
		}
	}

	return getProductPrices(bson.D{{"_id", bson.M{"$in": idsBSON}}}, 0, 0, bson.D{{"created_at", 1}})
}

func GetPricesForProduct(productID int, productType helpers.ProductType, cc steamapi.ProductCC) (prices []ProductPrice, err error) {

	var filter = bson.D{{"prod_cc", string(cc)}}

	if productType == helpers.ProductTypeApp {
		filter = append(filter, bson.E{Key: "app_id", Value: productID})
	} else if productType == helpers.ProductTypePackage {
		filter = append(filter, bson.E{Key: "package_id", Value: productID})
	} else {
		return prices, errors.New("invalid product type")
	}

	return getProductPrices(filter, 0, 0, bson.D{{"created_at", 1}})
}

func GetPricesForApps(appIDs []int, cc steamapi.ProductCC) (prices []ProductPrice, err error) {

	if len(appIDs) > 10 {
		appIDs = appIDs[0:10]
	}

	var a = bson.A{}
	for _, v := range appIDs {
		a = append(a, v)
	}

	var filter = bson.D{
		{Key: "prod_cc", Value: string(cc)},
		{Key: "app_id", Value: bson.M{"$in": a}},
	}

	return getProductPrices(filter, 0, 0, bson.D{{"created_at", 1}})
}

func GetPrices(offset int64, limit int64, filter bson.D) (prices []ProductPrice, err error) {

	return getProductPrices(filter, offset, limit, bson.D{{"created_at", -1}})
}

func getProductPrices(filter bson.D, offset int64, limit int64, sort bson.D) (prices []ProductPrice, err error) {

	cur, ctx, err := Find(CollectionProductPrices, offset, limit, sort, filter, nil, nil)
	if err != nil {
		return prices, err
	}

	defer func() {
		err = cur.Close(ctx)
		if err != nil {
			zap.S().Error(err)
		}
	}()

	for cur.Next(ctx) {

		var price ProductPrice
		err := cur.Decode(&price)
		if err != nil {
			zap.S().Error(err, price)
		} else {
			prices = append(prices, price)
		}
	}

	return prices, cur.Err()
}
