package queue

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/streadway/amqp"
)

func processApp(msg amqp.Delivery) (err error) {

	// Get message payload
	message := new(AppMessage)

	err = json.Unmarshal(msg.Body, message)
	if err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(msg.Body))
		}

		msg.Nack(false, false)
		return nil
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

	err = app.Fill()
	if err != nil {

		if strings.HasSuffix(err.Error(), "connect: connection refused") {
			time.Sleep(time.Second * 1)
			msg.Nack(false, true)
			return nil
		}

		logger.Error(err)
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
	return nil
}

type AppMessage struct {
	Time     time.Time
	AppID    int
	ChangeID int
}
