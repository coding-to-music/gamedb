package mongo

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Sale struct {
	SubID           int                     `bson:"sub_id"`
	SubOrder        int                     `bson:"sub_order"` // Order in the API response
	AppID           int                     `bson:"app_id"`
	AppName         string                  `bson:"app_name"`
	AppIcon         string                  `bson:"app_icon"`
	AppRating       float64                 `bson:"app_rating"`
	AppReleaseDate  time.Time               `bson:"app_date"`
	AppPrices       map[steam.ProductCC]int `bson:"app_prices"`
	AppLowestPrice  map[steam.ProductCC]int `bson:"app_lowest_price"`
	AppPlayersWeek  int                     `bson:"app_players"`
	AppCategories   []int                   `bson:"app_categories"`
	AppType         string                  `bson:"app_type"`
	AppPlatforms    []string                `bson:"app_platforms"`
	AppTags         []int                   `bson:"app_tags"`
	SaleStart       time.Time               `bson:"offer_start"`
	SaleEnd         time.Time               `bson:"offer_end"`
	SaleEndEstimate bool                    `bson:"offer_end_estimate"`
	SaleType        string                  `bson:"offer_type"`
	SalePercent     int                     `bson:"offer_percent"`
	SaleName        string                  `bson:"offer_name"`
}

func (offer Sale) BSON() bson.D {

	return bson.D{
		{"_id", offer.GetKey()},
		{"sub_id", offer.SubID},
		{"sub_order", offer.SubOrder},
		{"app_id", offer.AppID},
		{"app_name", offer.AppName},
		{"app_icon", offer.AppIcon},
		{"app_rating", offer.AppRating},
		{"app_date", offer.AppReleaseDate},
		{"app_prices", offer.AppPrices},
		{"app_lowest_price", offer.AppLowestPrice},
		{"app_players", offer.AppPlayersWeek},
		{"app_categories", offer.AppCategories},
		{"app_type", offer.AppType},
		{"app_platforms", offer.AppPlatforms},
		{"app_tags", offer.AppTags},
		{"offer_start", offer.SaleStart},
		{"offer_end", offer.SaleEnd},
		{"offer_end_estimate", offer.SaleEndEstimate},
		{"offer_type", offer.SaleType},
		{"offer_percent", offer.SalePercent},
		{"offer_name", offer.SaleName},
	}
}

func (offer Sale) GetKey() (ret string) {
	return strconv.Itoa(offer.AppID) + "-" + strconv.Itoa(offer.SubID)
}

func (offer Sale) GetType() string {
	return strings.Title(offer.SaleType)
}

func (offer Sale) GetOfferName() string {
	return strings.TrimPrefix(offer.SaleName, "Buy ")
}

func (offer Sale) GetAppRating() string {
	if offer.AppRating == 0 {
		return "-"
	} else {
		return helpers.FloatToString(offer.AppRating, 1) + "%"
	}
}

func (offer Sale) GetPriceInt() string {
	if offer.AppRating == 0 {
		return "-"
	} else {
		return helpers.FloatToString(offer.AppRating, 1) + "%"
	}
}

func (offer Sale) GetPriceString(code steam.ProductCC) string {

	priceInt, ok := offer.AppPrices[code]
	if ok {
		cc := helpers.GetProdCC(code)
		return helpers.FormatPrice(cc.CurrencyCode, priceInt)
	} else {
		return "-"
	}
}

// 0:not lowest - 1:match lowest - 2:lowest ever
func (offer Sale) IsLowest(code steam.ProductCC) int {

	price, ok := offer.AppPrices[code]
	if !ok {
		return 0
	}

	lowestPrice, ok := offer.AppLowestPrice[code]
	if !ok {
		return 0
	}

	if price == lowestPrice {
		return 1
	}

	if price < lowestPrice {
		return 2
	}

	return 0
}

func GetAppSales(appID int) (offers []Sale, err error) {
	return getSales(0, 0, bson.D{{"app_id", appID}}, bson.D{{"offer_end", 1}}, bson.M{"sub_id": 1, "offer_start": 1})
}

func CountSales() (count int64, err error) {

	var item = helpers.MemcacheSalesCount

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		return CountDocuments(CollectionAppSales, bson.D{{"offer_end", bson.M{"$gt": time.Now()}}}, 0)
	})

	return count, err
}

func GetHighestSaleOrder() (int, error) {

	sales, err := getSales(0, 1, bson.D{{"offer_end", bson.M{"$gt": time.Now()}}}, bson.D{{"sub_order", -1}}, bson.M{"sub_order": 1})
	if err != nil {
		return 0, err
	}
	return sales[0].SubOrder, nil
}

func GetAllSales(offset int64, limit int64, filter bson.D, sort bson.D) (offers []Sale, err error) {
	return getSales(offset, limit, filter, sort, nil)
}

func getSales(offset int64, limit int64, filter bson.D, sort bson.D, projection bson.M) (offers []Sale, err error) {

	cur, ctx, err := Find(CollectionAppSales, offset, limit, sort, filter, projection, nil)
	if err != nil {
		return offers, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var sale Sale
		err := cur.Decode(&sale)
		if err != nil {
			log.Err(err)
		}
		offers = append(offers, sale)
	}

	return offers, cur.Err()
}

func UpdateSales(offers []Sale) (err error) {

	if len(offers) < 1 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for _, offer := range offers {

		write := mongo.NewReplaceOneModel()
		write.SetFilter(bson.M{"_id": offer.GetKey()})
		write.SetReplacement(offer.BSON())
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	collection := client.Database(MongoDatabase).Collection(CollectionAppSales.String())
	_, err = collection.BulkWrite(ctx, writes, options.BulkWrite())
	return err
}

func GetUniqueSaleTypes() (types []string, err error) {

	var item = helpers.MemcacheUniqueSaleTypes

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &types, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return types, err
		}

		c := client.Database(MongoDatabase, options.Database()).Collection(CollectionAppSales.String())

		resp, err := c.Distinct(ctx, "offer_type", bson.M{"offer_end": bson.M{"$gt": time.Now()}}, options.Distinct())
		if err != nil {
			return types, err
		}

		for _, v := range resp {
			if val, ok := v.(string); ok {
				types = append(types, val)
			}
		}

		sort.Slice(types, func(i, j int) bool {
			return types[i] < types[j]
		})

		return types, err
	})

	return types, err
}
