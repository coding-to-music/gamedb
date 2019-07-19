package mongo

import (
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steam"
)

type Offer struct {
	CreatedAt      time.Time               `bson:"created_at"`
	UpdatedAt      time.Time               `bson:"updated_at"`
	SubID          int                     `bson:"sub_id"`
	SubOrder       int                     `bson:"sub_order"`
	AppID          int                     `bson:"app_id"`
	AppRating      int                     `bson:"app_rating"`
	AppReleaseDate time.Time               `bson:"app_date"`
	AppPrices      map[steam.ProductCC]int `bson:"app_prices"`
	AppLowestPrice map[steam.ProductCC]int `bson:"app_lowest_price"`
	AppPlayersWeek int                     `bson:"app_players"`
	OfferStart     time.Time               `bson:"offer_start"`
	OfferEnd       time.Time               `bson:"offer_end"`
	OfferType      string                  `bson:"offer_type"`
	OfferPercent   int                     `bson:"offer_percent"`
}

func (offer Offer) BSON() (ret interface{}) {

	if offer.CreatedAt.IsZero() {
		offer.CreatedAt = time.Now()
	}
	offer.UpdatedAt = time.Now()

	return M{
		"_id":        offer.getKey(),
		"created_at": offer.CreatedAt,
		"updated_at": offer.UpdatedAt,
		"app_id":     offer.AppID,
		"sub_id":     offer.SubID,
		"ends":       offer.OfferEnd,
		"type":       offer.OfferType,
	}
}

func (offer Offer) getKey() (ret interface{}) {
	return strconv.Itoa(offer.AppID) + "-" + strconv.Itoa(offer.SubID)
}
