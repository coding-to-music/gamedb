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

func processPackage(msg amqp.Delivery) (err error) {

	return

	// Get message
	message := new(PackageMessage)

	err = json.Unmarshal(msg.Body, message)
	if err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(msg.Body))
		}

		msg.Nack(false, false)
		return nil
	}

	// Update package
	db, err := mysql.GetDB()
	if err != nil {
		logger.Error(err)
	}

	pack := new(mysql.Package)

	db.Attrs(mysql.GetDefaultPackageJSON()).FirstOrCreate(pack, mysql.Package{ID: message.PackageID})

	if message.ChangeID != 0 {
		pack.ChangeID = message.ChangeID
	}

	priceBeforeFill := pack.PriceFinal

	// Move all the stuff in here to queue?
	pack.Fill()

	db.Save(pack)
	if db.Error != nil {
		logger.Error(db.Error)
	}

	// Save price change
	price := new(datastore.PackagePrice)
	price.CreatedAt = time.Now()
	price.PackageID = pack.ID
	price.PriceInitial = pack.PriceInitial
	price.PriceFinal = pack.PriceFinal
	price.Discount = pack.PriceDiscount
	price.Currency = "usd"
	price.Change = pack.PriceFinal - priceBeforeFill

	if price.Change != 0 {
		_, err = datastore.SaveKind(price.GetKey(), price)
		if err != nil {
			logger.Error(err)
		}
	}

	msg.Ack(false)
	return nil
}

type PackageMessage struct {
	PackageID int
	ChangeID  int
}
