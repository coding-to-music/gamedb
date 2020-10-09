package queue

import (
	"sync"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/steam"
	"github.com/gamedb/gamedb/pkg/websockets"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type PackagePriceMessage struct {
	PackageID   uint               `json:"id"`
	PackageName string             `json:"package_name"`
	PackageIcon string             `json:"package_icon"`
	ProductCC   steamapi.ProductCC `json:"prod_cc"`
	Time        time.Time          `json:"time"`
	BeforePrice *int               `json:"before_price"`
	LowestPrice *int               `json:"lowest_price"`
}

func packagePriceHandler(message *rabbit.Message) {

	payload := PackagePriceMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	var productCC = i18n.GetProdCC(payload.ProductCC)

	// Get package details
	response, err := steam.GetSteam().GetPackageDetails(payload.PackageID, productCC.ProductCode, steamapi.LanguageEnglish)
	err = steam.AllowSteamCodes(err)
	if err == steamapi.ErrPackageNotFound {
		message.Ack()
		return
	}
	if err != nil {
		steam.LogSteamError(err)
		sendToRetryQueue(message)
		return
	}

	var wg sync.WaitGroup

	// Update package price
	wg.Add(1)
	go func() {

		defer wg.Done()

		var update = bson.D{
			{"prices." + string(productCC.ProductCode),
				helpers.ProductPrice{
					Currency:        response.Data.Price.Currency,
					Initial:         response.Data.Price.Initial,
					Final:           response.Data.Price.Final,
					DiscountPercent: response.Data.Price.DiscountPercent,
					Individual:      response.Data.Price.Individual,
				},
			},
		}

		_, err = mongo.UpdateOne(mongo.CollectionPackages, bson.D{{"_id", payload.PackageID}}, update)
		if err != nil {
			log.ErrS(err)
			sendToRetryQueue(message)
		}
	}()

	if payload.BeforePrice != nil {

		// Save price change
		var oldPrice = *payload.BeforePrice
		var newPrice = response.Data.Price.Final

		wg.Add(1)
		go func() {

			defer wg.Done()

			if payload.BeforePrice != nil {

				price := mongo.ProductPrice{}
				price.PackageID = int(payload.PackageID)
				price.Name = payload.PackageName
				price.Icon = payload.PackageIcon
				price.CreatedAt = time.Now()
				price.Currency = productCC.CurrencyCode
				price.ProdCC = productCC.ProductCode
				price.PriceBefore = oldPrice
				price.PriceAfter = newPrice
				price.Difference = newPrice - oldPrice
				if oldPrice > 0 {
					price.DifferencePercent = (float64(newPrice-oldPrice) / float64(oldPrice)) * 100
				}

				result, err := mongo.InsertOne(mongo.CollectionProductPrices, price)
				if err != nil {
					log.ErrS(err)
					return
				}

				// Send websockets to prices page
				if result != nil {
					if insertedID, ok := result.InsertedID.(primitive.ObjectID); ok {

						wsPayload := StringsPayload{IDs: []string{insertedID.Hex()}}
						err2 := ProduceWebsocket(wsPayload, websockets.PagePrices)
						if err2 != nil {
							log.ErrS(err2)
						}
					}
				}
			}
		}()

		// Post to Discord
		// wg.Add(1)
		// go func() {
		//
		// 	defer wg.Done()
		//
		// 	if productCC.ProductCode == steamapi.ProductCCUS &&
		// 		oldPrice > newPrice && // Incase it goes from -90% to -80%
		// 		newPrice > 0 { // Free games are usually just removed from the store
		//
		// 		var msg = "Package " + strconv.FormatUint(uint64(payload.PackageID), 10) + ": " + helpers.GetPackageName(int(payload.PackageID), payload.PackageName)
		// 		_, err := discordClient.ChannelMessageSend("685246060930924544", msg)
		// 		if err != nil {
		// 			log.ErrS(err)
		// 		}
		// 	}
		// }()

	}

	wg.Wait()

	//
	message.Ack()
}
