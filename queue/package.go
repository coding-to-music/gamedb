package queue

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/streadway/amqp"
)

func processPackage(msg amqp.Delivery) (ack bool, requeue bool, err error) {

	// Get message
	message := new(RabbitMessagePackage)

	err = json.Unmarshal(msg.Body, message)
	if err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(msg.Body))
		}

		return false, false, err
	}

	// Read PICS JSON
	var root = map[string]string{}
	var extended = map[string]string{}
	var appids []string
	var depotids []string
	var appitems []string

	// Base
	root["name"] = message.KeyValues.Name

	for _, v := range message.KeyValues.Children {
		if v.Value != nil {
			root[v.Name] = v.Value.(string)
		} else if v.Name == "extended" {
			// Extended
			for _, vv := range v.Children {
				extended[vv.Name] = vv.Value.(string)
			}
		} else if v.Name == "appids" {
			// App IDs
			for _, vv := range v.Children {
				appids = append(appids, vv.Value.(string))
			}
		} else if v.Name == "depotids" {
			// Depot IDs
			for _, vv := range v.Children {
				depotids = append(depotids, vv.Value.(string))
			}
		} else if v.Name == "appitems" {
			// App Items
			for _, vv := range v.Children {
				appitems = append(appitems, vv.Value.(string))
			}
		} else {
			fmt.Printf("Package %s has a '%s' section", root["packageid"], v.Name)
		}
	}

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

	loop(flatMap, "", message.KeyValues, true)

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

type RabbitMessagePackage struct {
	ID           int                           `json:"ID"`
	ChangeNumber int                           `json:"ChangeNumber"`
	MissingToken bool                          `json:"MissingToken"`
	SHAHash      string                        `json:"SHAHash"`
	KeyValues    RabbitMessagePackageKeyValues `json:"KeyValues"`
	OnlyPublic   bool                          `json:"OnlyPublic"`
	UseHTTP      bool                          `json:"UseHttp"`
	HTTPURI      interface{}                   `json:"HttpUri"`
}

type RabbitMessagePackageKeyValues struct {
	Name     string                          `json:"Name"`
	Value    interface{}                     `json:"Value"`
	Children []RabbitMessagePackageKeyValues `json:"Children"`
}
