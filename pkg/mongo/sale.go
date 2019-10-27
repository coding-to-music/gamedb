package mongo

import (
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/log"
	. "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Sale struct {
	SubID            int                     `bson:"sub_id"`
	SubOrder         int                     `bson:"sub_order"` // Order in the API response
	AppID            int                     `bson:"app_id"`
	AppName          string                  `bson:"app_name"`
	AppIcon          string                  `bson:"app_icon"`
	AppRating        float64                 `bson:"app_rating"`
	AppReleaseDate   time.Time               `bson:"app_date"`
	AppPrices        map[steam.ProductCC]int `bson:"app_prices"`
	AppLowestPrice   map[steam.ProductCC]int `bson:"app_lowest_price"`
	AppPlayersWeek   int                     `bson:"app_players"`
	OfferStart       time.Time               `bson:"offer_start"`
	OfferEnd         time.Time               `bson:"offer_end"`
	OfferEndEstimate bool                    `bson:"offer_end_estimate"`
	OfferType        string                  `bson:"offer_type"`
	OfferPercent     int                     `bson:"offer_percent"`
}

func (offer Sale) BSON() (ret interface{}) {

	return M{
		"_id":                offer.getKey(),
		"sub_id":             offer.SubID,
		"sub_order":          offer.SubOrder,
		"app_id":             offer.AppID,
		"app_name":           offer.AppName,
		"app_icon":           offer.AppIcon,
		"app_rating":         offer.AppRating,
		"app_date":           offer.AppReleaseDate,
		"app_prices":         offer.AppPrices,
		"app_lowest_price":   offer.AppLowestPrice,
		"app_players":        offer.AppPlayersWeek,
		"offer_start":        offer.OfferStart,
		"offer_end":          offer.OfferEnd,
		"offer_end_estimate": offer.OfferEndEstimate,
		"offer_type":         offer.OfferType,
		"offer_percent":      offer.OfferPercent,
	}
}

func (offer Sale) getKey() (ret string) {
	return strconv.Itoa(offer.AppID) + "-" + strconv.Itoa(offer.SubID)
}

func GetAppOffers(appID int) (offers []Sale, err error) {
	return getOffers(0, 0, D{{"app_id", appID}}, M{"sub_id": 1})
}

func GetAllOffers(offset int64, limit int64, filter D) (offers []Sale, err error) {
	return getOffers(offset, limit, filter, nil)
}

func getOffers(offset int64, limit int64, filter D, projection M) (offers []Sale, err error) {

	var sort = D{{"offer_end", 1}}

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
		write.SetFilter(M{"_id": offer.getKey()})
		write.SetReplacement(offer.BSON())
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	collection := client.Database(MongoDatabase).Collection(CollectionAppSales.String())
	_, err = collection.BulkWrite(ctx, writes, options.BulkWrite())
	return err
}
