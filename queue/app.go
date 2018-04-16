package queue

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/steam"
	"github.com/streadway/amqp"
)

func processApp(msg amqp.Delivery) {

	// Get message payload
	message := new(AppMessage)

	err := json.Unmarshal(msg.Body, message)
	if err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(msg.Body))
		}

		msg.Nack(false, false)
		return
	}

	// Get news
	_, err = datastore.GetArticlesFromSteam(message.AppID)
	if err != nil {
		logger.Error(err)
	}

	// Update app
	db, err := mysql.GetDB()
	if err != nil {
		logger.Error(err)
	}

	app := new(mysql.App)

	db.Attrs(mysql.GetDefaultAppJSON()).FirstOrCreate(app, mysql.App{ID: message.AppID})

	if message.ChangeID != 0 {
		app.ChangeNumber = message.ChangeID
	}

	priceBeforeFill := app.PriceFinal

	errs := app.Update()
	if len(errs) > 0 {

		for _, v := range errs {
			logger.Error(v)
		}

		// API is probably down
		for _, v := range errs {
			if v.Error() == steam.ErrInvalidJson {
				time.Sleep(time.Second * 10)
				msg.Nack(false, true)
				return
			}
		}

		for _, v := range errs {
			if strings.HasSuffix(v.Error(), "connect: connection refused") {
				time.Sleep(time.Second * 10)
				msg.Nack(false, true)
				return
			}
		}
	}

	db.Save(app)
	if db.Error != nil {
		logger.Error(err)
	}

	// Save price change
	price := new(datastore.Price)
	price.CreatedAt = time.Now()
	price.AppID = app.ID
	price.Name = app.GetName()
	price.PriceInitial = app.PriceInitial
	price.PriceFinal = app.PriceFinal
	price.Discount = app.PriceDiscount
	price.Currency = "usd"
	price.Change = app.PriceFinal - priceBeforeFill
	price.Icon = app.Icon
	price.ReleaseDateNice = app.GetReleaseDateNice()
	price.ReleaseDateUnix = app.GetReleaseDateUnix()

	if price.Change != 0 {
		_, err = datastore.SaveKind(price.GetKey(), price)
		if err != nil {
			logger.Error(err)
		}
	}

	// Ack
	msg.Ack(false)
	return
}

type AppMessage struct {
	Time     time.Time
	AppID    int
	ChangeID int
}
