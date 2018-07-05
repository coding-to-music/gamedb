package queue

import (
	"encoding/json"
	"strings"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/streadway/amqp"
)

type RabbitMessageApp struct {
	baseQueue
	RabbitMessageProduct
}

func (d RabbitMessageApp) getQueueName() string {
	return QueueAppsData
}

func (d *RabbitMessageApp) process(msg amqp.Delivery) (ack bool, requeue bool) {

	// Get message payload
	message := new(RabbitMessageApp)

	err := json.Unmarshal(msg.Body, message)
	if err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(msg.Body))
		}

		return false, false
	}

	return false, true

	//app := new(mysql.App)
	//
	//// Update app
	//db, err := mysql.GetDB()
	//if err != nil {
	//	logger.Error(err)
	//}
	//
	//db.Attrs(mysql.GetDefaultAppJSON()).FirstOrCreate(app, mysql.App{ID: message.AppID})
	//if db.Error != nil {
	//	logger.Error(db.Error)
	//}
	//
	//if message.PICSChangeID != 0 {
	//	app.ChangeNumber = message.PICSChangeID
	//}
	//
	//priceBeforeFill := app.PriceFinal
	//
	//errs := app.UpdateFromAPI()
	//if len(errs) > 0 {
	//	// Nack on hard fails
	//	for _, err = range errs {
	//		if err, ok := err.(mysql.UpdateError); ok {
	//			if err.IsHard() {
	//				return false, false
	//			}
	//		}
	//	}
	//	// Retry on all other errors
	//	for _, err = range errs {
	//		if err != steam.ErrNullResponse {
	//			logger.Error(err)
	//		}
	//		return false, true
	//	}
	//}
	////if v.Error() == steam.ErrInvalidJson || v == steam.ErrBadResponse || strings.HasSuffix(v.Error(), "connect: connection refused") {
	////	return false, true
	////}
	//
	//db.Save(app)
	//if db.Error != nil {
	//	logger.Error(db.Error)
	//}
	//
	//// Save price change
	//price := new(datastore.Price)
	//price.CreatedAt = time.Now()
	//price.AppID = app.ID
	//price.PICSName = app.GetName()
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
	//	prices, err := datastore.GetAppPrices(app.ID, 1)
	//	if err != nil {
	//		logger.Error(err)
	//	}
	//
	//	if len(prices) == 0 {
	//		price.First = true
	//	}
	//
	//	_, err = datastore.SaveKind(price.GetKey(), price)
	//	if err != nil {
	//		logger.Error(err)
	//	}
	//}
	//
	//return true, false
}
