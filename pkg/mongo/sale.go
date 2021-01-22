package mongo

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Sale struct {
	SubID                int                        `bson:"sub_id"`
	SubOrder             int                        `bson:"sub_order"` // Order in the API response
	AppID                int                        `bson:"app_id"`
	AppName              string                     `bson:"app_name"`
	AppIcon              string                     `bson:"app_icon"`
	AppRating            float64                    `bson:"app_rating"`
	AppReleaseDate       time.Time                  `bson:"app_date"`
	AppReleaseDateString string                     `bson:"app_date_string"`
	AppPrices            map[steamapi.ProductCC]int `bson:"app_prices"`
	AppLowestPrice       map[steamapi.ProductCC]int `bson:"app_lowest_price"`
	AppPlayersWeek       int                        `bson:"app_players"`
	AppCategories        []int                      `bson:"app_categories"`
	AppType              string                     `bson:"app_type"`
	AppPlatforms         []string                   `bson:"app_platforms"`
	AppTags              []int                      `bson:"app_tags"`
	SaleStart            time.Time                  `bson:"offer_start"`
	SaleEnd              time.Time                  `bson:"offer_end"`
	SaleEndEstimate      bool                       `bson:"offer_end_estimate"`
	SaleType             string                     `bson:"offer_type"`
	SalePercent          int                        `bson:"offer_percent"`
	SaleName             string                     `bson:"offer_name"`
}

func (sale Sale) BSON() bson.D {

	return bson.D{
		{"_id", sale.GetKey()},
		{"sub_id", sale.SubID},
		{"sub_order", sale.SubOrder},
		{"app_id", sale.AppID},
		{"app_name", sale.AppName},
		{"app_icon", sale.AppIcon},
		{"app_rating", sale.AppRating},
		{"app_date", sale.AppReleaseDate},
		{"app_prices", sale.AppPrices},
		{"app_lowest_price", sale.AppLowestPrice},
		{"app_players", sale.AppPlayersWeek},
		{"app_categories", sale.AppCategories},
		{"app_type", sale.AppType},
		{"app_platforms", sale.AppPlatforms},
		{"app_tags", sale.AppTags},
		{"offer_start", sale.SaleStart},
		{"offer_end", sale.SaleEnd},
		{"offer_end_estimate", sale.SaleEndEstimate},
		{"offer_type", sale.SaleType},
		{"offer_percent", sale.SalePercent},
		{"offer_name", sale.SaleName},
	}
}

func ensureSaleIndexes() {

	var indexModels []mongo.IndexModel

	cols := []string{
		"app_date",
		"app_rating",
		"offer_end",
		"offer_name",
		"offer_percent",
	}

	// Price fields
	for _, v := range i18n.GetProdCCs(true) {
		cols = append(cols, "app_prices."+string(v.ProductCode))
		cols = append(cols, "app_lowest_price."+string(v.ProductCode))
	}

	//
	for _, col := range cols {
		indexModels = append(indexModels, mongo.IndexModel{
			Keys: bson.D{{col, 1}},
		}, mongo.IndexModel{
			Keys: bson.D{{col, -1}},
		})
	}

	//
	client, ctx, err := getMongo()
	if err != nil {
		log.ErrS(err)
		return
	}

	_, err = client.Database(config.C.MongoDatabase).Collection(CollectionAppSales.String()).Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		log.ErrS(err)
	}
}

func (sale Sale) GetKey() (ret string) {
	return strconv.Itoa(sale.AppID) + "-" + strconv.Itoa(sale.SubID)
}

func (sale Sale) GetType() string {
	return strings.Title(sale.SaleType)
}

func (sale Sale) GetOfferName() string {
	return strings.TrimPrefix(sale.SaleName, "Buy ")
}

func (sale Sale) GetAppRating() string {

	if sale.AppRating == 0 {
		return "-"
	}
	return helpers.FloatToString(sale.AppRating, 1) + "%"
}

func (sale Sale) GetPriceInt() string {

	if sale.AppRating == 0 {
		return "-"
	}
	return helpers.FloatToString(sale.AppRating, 1) + "%"
}

func (sale Sale) GetPriceString(code steamapi.ProductCC) string {

	priceInt, ok := sale.AppPrices[code]
	if ok {
		cc := i18n.GetProdCC(code)
		return i18n.FormatPrice(cc.CurrencyCode, priceInt)
	}
	return "-"
}

// 0:not lowest - 1:match lowest - 2:lowest ever
func (sale Sale) IsLowest(code steamapi.ProductCC) int {

	price, ok := sale.AppPrices[code]
	if !ok {
		return 0
	}

	lowestPrice, ok := sale.AppLowestPrice[code]
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

	cur, ctx, err := find(CollectionAppSales, offset, limit, sort, filter, projection, nil)
	if err != nil {
		return offers, err
	}

	defer closeCursor(cur, ctx)

	for cur.Next(ctx) {

		var sale Sale
		err := cur.Decode(&sale)
		if err != nil {
			log.ErrS(err)
		} else {
			offers = append(offers, sale)
		}
	}

	return offers, cur.Err()
}

func ReplaceSales(offers []Sale) (err error) {

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

	collection := client.Database(config.C.MongoDatabase).Collection(CollectionAppSales.String())
	_, err = collection.BulkWrite(ctx, writes, options.BulkWrite())
	return err
}

func GetUniqueSaleTypes() (types []string, err error) {

	err = memcache.GetSetInterface(memcache.ItemUniqueSaleTypes, &types, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return types, err
		}

		c := client.Database(config.C.MongoDatabase, options.Database()).Collection(CollectionAppSales.String())

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
