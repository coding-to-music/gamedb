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

func processPackage(msg amqp.Delivery) (ack bool, requeue bool) {

	// Get message
	message := new(PackageMessage)

	err := json.Unmarshal(msg.Body, message)
	if err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(msg.Body))
		}

		return false, false
	}

	// Update package
	db, err := mysql.GetDB()
	if err != nil {
		logger.Error(err)
	}

	pack := new(mysql.Package)

	db.Attrs(mysql.GetDefaultPackageJSON()).FirstOrCreate(pack, mysql.Package{ID: message.PackageID})
	if db.Error != nil {
		logger.Error(db.Error)
	}

	if message.ChangeID != 0 {
		pack.ChangeID = message.ChangeID
	}

	priceBeforeFill := pack.PriceFinal

	errs := pack.Update()
	if len(errs) > 0 {

		for _, v := range errs {
			logger.Error(v)
		}

		for _, v := range errs {
			if v.Error() == steam.ErrInvalidJson || v == steam.ErrBadResponse || strings.HasSuffix(v.Error(), "connect: connection refused") {
				return false, true
			}
		}
	}

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

	return true, false
}

type PackageMessage struct {
	Time      time.Time
	PackageID int
	ChangeID  int
}
