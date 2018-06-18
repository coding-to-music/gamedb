package queue

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/streadway/amqp"
)

func processPackage(msg amqp.Delivery) (ack bool, requeue bool, err error) {

	// Get message
	message := new(RabbitMessageProduct)

	err = json.Unmarshal(msg.Body, message)
	if err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(msg.Body))
		}

		return false, false, err
	}

	x, _ := json.Marshal(message.KeyValues)
	y, _ := json.Marshal(message.KeyValues.Convert())

	// Update package
	db, err := mysql.GetDB()
	if err != nil {
		logger.Error(err)
	}

	pack := new(mysql.Package)

	// Save raw data
	bytes, err := json.Marshal(message.KeyValues)
	if err != nil {
		println(err.Error())
	}
	pack.RawPICS = string(bytes)

	flatMap := map[string]interface{}{}

	//loop(flatMap, "", message.KeyValues, true)

	for k, v := range flatMap {
		println(k + ": " + v.(string))
	}

	println(" ")

	return false, true, nil

	//db.Attrs(mysql.GetDefaultPackageJSON()).FirstOrCreate(pack, mysql.Package{ID: message.PackageID})
	//if db.Error != nil {
	//	logger.Error(db.Error)
	//}
	//
	//if message.ChangeID != 0 {
	//	pack.ChangeID = message.ChangeID
	//}

	priceBeforeFill := pack.PriceFinal

	errs := pack.Update()
	if len(errs) > 0 {
		// Nack on hard fails
		for _, err = range errs {
			if err, ok := err.(mysql.UpdateError); ok {
				if err.IsHard() {
					return false, false, err
				}
			}
		}
		// Retry on all other errors
		for _, err = range errs {
			logger.Error(err)
			return false, true, err
		}
	}
	//if v.Error() == steam.ErrInvalidJson || v == steam.ErrNullResponse || strings.HasSuffix(v.Error(), "connect: connection refused") {
	//	return false, true
	//}

	db.Save(pack)
	if db.Error != nil {
		logger.Error(db.Error)
	}

	// Save price change
	price := new(datastore.Price)
	price.CreatedAt = time.Now()
	price.PackageID = pack.ID
	price.Name = pack.GetName()
	price.PriceInitial = pack.PriceInitial
	price.PriceFinal = pack.PriceFinal
	price.Discount = pack.PriceDiscount
	price.Currency = "usd"
	price.Change = pack.PriceFinal - priceBeforeFill
	price.Icon = pack.GetDefaultAvatar()
	price.ReleaseDateNice = pack.GetReleaseDateNice()
	price.ReleaseDateUnix = pack.GetReleaseDateUnix()

	if price.Change != 0 {

		prices, err := datastore.GetPackagePrices(pack.ID, 1)
		if err != nil {
			logger.Error(err)
		}

		if len(prices) == 0 {
			price.First = true
		}

		_, err = datastore.SaveKind(price.GetKey(), price)
		if err != nil {
			logger.Error(err)
		}
	}

	return true, false, nil
}
