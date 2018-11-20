package queue

import (
	"github.com/gamedb/website/helpers"
	"github.com/streadway/amqp"
)

type RabbitMessageApp struct {
	RabbitMessageProduct
}

func (d RabbitMessageApp) getQueueName() RabbitQueue {
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
	message := new(RabbitMessageApp)

	err = helpers.Unmarshal(msg.Body, message)
	if err != nil {
		return false, false, err
	}

	return false, true, err

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
	//	app.ChangeNumber = message.PICSChangeID
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
