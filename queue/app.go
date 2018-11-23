package queue

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/logging"
	"github.com/gamedb/website/websockets"
	"github.com/streadway/amqp"
)

type RabbitMessageApp struct {
	PICSAppInfo RabbitMessageProduct
}

func (d RabbitMessageApp) getConsumeQueue() RabbitQueue {
	return QueueAppsData
}

func (d RabbitMessageApp) getProduceQueue() RabbitQueue {
	return QueueApps
}

func (d RabbitMessageApp) getRetryData() RabbitMessageDelay {
	return RabbitMessageDelay{}
}

func (d RabbitMessageApp) process(msg amqp.Delivery) (ack bool, requeue bool, err error) {

	// Get message payload
	rabbitMessage := new(RabbitMessageApp)

	err = helpers.Unmarshal(msg.Body, rabbitMessage)
	if err != nil {
		return false, false, err
	}

	message := rabbitMessage.PICSAppInfo

	logging.Info("Consuming app: " + strconv.Itoa(message.ID))

	if !db.IsValidAppID(message.ID) {
		return false, false, errors.New("invalid app ID: " + strconv.Itoa(message.ID))
	}

	// Load current app
	gorm, err := db.GetMySQLClient()
	if err != nil {
		return false, true, err
	}

	app := db.App{}
	gorm.First(&app, message.ID)
	if gorm.Error != nil && !gorm.RecordNotFound() {
		return false, true, gorm.Error
	}

	if app.PICSChangeNumber >= message.ChangeNumber {
		return true, false, nil
	}

	var appBeforeUpdate = app

	// Update with new details
	app.ID = message.ID
	app.PICSChangeNumber = message.ChangeNumber
	app.Name = message.KeyValues.Name
	app.PICSRaw = string(msg.Body)

	for _, v := range message.KeyValues.Children {

		switch v.Name {
		case "appid":
			var i64 int64
			i64, err = strconv.ParseInt(v.Value.(string), 10, 32)
			app.ID = int(i64)
		case "common":

			var common = db.PICSAppCommon{}
			for _, vv := range v.Children {
				if vv.Value == nil {
					bytes, err := json.Marshal(vv.GetChildrenAsSlice()) // todo, flatten, not slice
					logging.Error(err)
					common[vv.Name] = string(bytes)
				} else {
					common[vv.Name] = vv.Value.(string)
				}
			}
			err = app.SetCommon(common)
			logging.Error(err)

		case "extended":

			var extended = db.PICSExtended{}
			for _, vv := range v.Children {
				if vv.Value == nil {
					bytes, err := json.Marshal(vv.GetChildrenAsSlice()) // todo, flatten, not slice
					logging.Error(err)
					extended[vv.Name] = string(bytes)
				} else {
					extended[vv.Name] = vv.Value.(string)
				}
			}
			err = app.SetExtended(extended)
			logging.Error(err)

		case "config":

			var config = db.PICSAppConfig{}
			for _, vv := range v.Children {
				if vv.Value == nil {
					bytes, err := json.Marshal(vv.GetChildrenAsSlice()) // todo, flatten, not slice
					logging.Error(err)
					config[vv.Name] = string(bytes)
				} else {
					config[vv.Name] = vv.Value.(string)
				}
			}
			err = app.SetConfig(config)
			logging.Error(err)

		case "depots":

			var depots = db.PICSAppDepots{}
			for _, vv := range v.Children {
				if vv.Value == nil {
					bytes, err := json.Marshal(vv.GetChildrenAsSlice()) // todo, flatten, not slice
					logging.Error(err)
					depots[vv.Name] = string(bytes)
				} else {
					depots[vv.Name] = vv.Value.(string)
				}
			}
			err = app.SetDepots(depots)
			logging.Error(err)

		default:
			logging.Info(v.Name + " field in PICS ignored (Change " + strconv.Itoa(app.PICSChangeNumber) + ")")
		}

		logging.Error(err)
	}

	// Update from API
	errs := app.UpdateFromAPI()
	for _, v := range errs {
		logging.Error(v)
	}
	for _, v := range errs {
		if v != steam.ErrAppNotFound {
			return false, true, err
		}
	}

	// Save new data
	gorm.Save(&app)
	if gorm.Error != nil {
		return false, true, gorm.Error
	}

	// Save price changes
	var prices db.ProductPrices
	var price db.ProductPriceCache
	var kinds []db.Kind
	for code := range steam.Countries {

		var oldPrice, newPrice int

		prices, err = appBeforeUpdate.GetPrices()
		if err == nil {
			price, err = prices.Get(code)
			if err == nil {
				oldPrice = price.Final
			} else {
				continue // Only compare if there is an old price to compare to
			}
		}

		prices, err = app.GetPrices()
		if err == nil {
			price, err = prices.Get(code)
			if err == nil {
				newPrice = price.Final
			}
		}

		if oldPrice != newPrice {
			kinds = append(kinds, db.CreateProductPrice(app, code, oldPrice, newPrice))
		}
	}

	err = db.BulkSaveKinds(kinds, db.KindProductPrice, true)
	if err != nil {
		return false, true, err
	}

	// Send websocket
	page, err := websockets.GetPage(websockets.PageApps)
	if err == nil && page.HasConnections() {
		//page.Send(app.OutputForJSON())
	}

	return true, false, err

	//app := new(db.App)
	//
	//// Update app
	//gorm, err := db.GetMySQLClient()
	//if err != nil {
	//	logging.Error(err)
	//}
	//
	//gorm.FirstOrCreate(app, db.App{ID: message.AppID})
	//if gorm.Error != nil {
	//	logging.Error(gorm.Error)
	//}
	//
	//if message.PICSChangeID != 0 {
	//	app.PICSChangeNumber = message.PICSChangeID
	//  app.ChangeNumberDate = time.now().Unix()
	//}
	//
	//priceBeforeFill := app.PriceFinal
	//
	//errs := app.UpdateFromAPI()
	//if len(errs) > 0 {
	//	// Nack on hard fails
	//	for _, err = range errs {
	//		if err2, ok := err.(gorm.UpdateError); ok {
	//			if err2.IsHard() {
	//				return false, false, err2
	//			}
	//		}
	//	}
	//	// Retry on all other errors
	//	for _, err = range errs {
	//		if err != steam.ErrNullResponse {
	//			logging.Error(err)
	//		}
	//		return false, true, err
	//	}
	//}
	////if v.Error() == steam.ErrInvalidJson || v == steam.ErrBadResponse || strings.HasSuffix(v.Error(), "connect: connection refused") {
	////	return false, true
	////}
	//
	//gorm.Save(app)
	//if gorm.Error != nil {
	//	logging.Error(gorm.Error)
	//}
	//
	//// Save price change
	//price := new(db.AppPrice)
	//price.CreatedAt = time.Now()
	//price.AppID = app.ID
	//price.Name = app.GetName()
	//price.PriceInitial = app.PriceInitial
	//price.PriceFinal = app.PriceFinal
	//price.Discount = app.PriceDiscount
	//price.Currency = "usd"
	//price.Change = app.PriceFinal - priceBeforeFill
	//price.Icon = app.Icon
	//price.ReleaseDateNice = app.GetReleaseDateNice()
	//price.ReleaseDateUnix = app.GetReleaseDateUnix()
	//
	//if price.Change != 0 {
	//
	//	prices, err := db.GetAppPrices(app.ID, 1)
	//	if err != nil {
	//		logging.Error(err)
	//	}
	//
	//	if len(prices) == 0 {
	//		price.First = true
	//	}
	//
	//	_, err = db.SaveKind(price.GetKey(), price)
	//	if err != nil {
	//		logging.Error(err)
	//	}
	//}
	//
	//return true, false, err
}
