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

type RabbitMessagePackage struct {
	PICSPackageInfo RabbitMessageProduct
}

func (d RabbitMessagePackage) getConsumeQueue() RabbitQueue {
	return QueuePackagesData
}

func (d RabbitMessagePackage) getProduceQueue() RabbitQueue {
	return QueuePackages
}

func (d RabbitMessagePackage) getRetryData() RabbitMessageDelay {
	return RabbitMessageDelay{}
}

func (d RabbitMessagePackage) process(msg amqp.Delivery) (ack bool, requeue bool, err error) {

	// Get message
	rabbitMessage := new(RabbitMessagePackage)

	err = helpers.Unmarshal(msg.Body, rabbitMessage)
	if err != nil {
		return false, false, err
	}

	message := rabbitMessage.PICSPackageInfo

	logging.Info("Consuming package: " + strconv.Itoa(message.ID))

	if !db.IsValidPackageID(message.ID) {
		return false, false, errors.New("invalid package ID")
	}

	// Load current package
	gorm, err := db.GetMySQLClient()
	if err != nil {
		return false, true, err
	}

	pack := new(db.Package)
	gorm.First(&pack, message.ID)
	if gorm.Error != nil && !gorm.RecordNotFound() {
		return false, true, gorm.Error
	}

	if pack.PICSChangeID >= message.ChangeNumber {
		return true, false, nil
	}

	var packageBeforeUpdate = pack

	// Update with new details
	pack.ID = message.ID
	pack.PICSChangeID = message.ChangeNumber
	pack.PICSName = message.KeyValues.Name
	pack.PICSRaw = string(msg.Body)

	for _, v := range message.KeyValues.Children {

		switch v.Name {
		case "billingtype":
			var i64 int64
			i64, err = strconv.ParseInt(v.Value.(string), 10, 8)
			pack.PICSBillingType = int8(i64)
		case "licensetype":
			var i64 int64
			i64, err = strconv.ParseInt(v.Value.(string), 10, 8)
			pack.PICSLicenseType = int8(i64)
		case "status":
			var i64 int64
			i64, err = strconv.ParseInt(v.Value.(string), 10, 8)
			pack.PICSStatus = int8(i64)
		case "packageid":
			// Empty
		case "appids":

			var i int
			var appIDs []int
			for _, vv := range v.Children {
				i, err = strconv.Atoi(vv.Value.(string))
				logging.Error(err)
				appIDs = append(appIDs, i)
			}
			err = pack.SetAppIDs(appIDs)
			logging.Error(err)

		case "depotids":

			var i int
			var depotIDs []int
			for _, vv := range v.Children {
				i, err = strconv.Atoi(vv.Value.(string))
				logging.Error(err)
				depotIDs = append(depotIDs, i)
			}
			err = pack.SetDepotIDs(depotIDs)
			logging.Error(err)

		case "appitems":

			var appItems = map[string]string{}
			for _, vv := range v.Children {
				if len(vv.Children) == 1 {
					appItems[vv.Name] = vv.Children[0].Value.(string)
				}
			}
			err = pack.SetAppItems(appItems)
			logging.Error(err)

		case "extended":

			var extended = db.Extended{}
			for _, vv := range v.Children {
				if vv.Value == nil {
					bytes, err := json.Marshal(vv.GetChildrenAsSlice())
					logging.Error(err)
					extended[vv.Name] = string(bytes)
				} else {
					extended[vv.Name] = vv.Value.(string)
				}
			}
			err = pack.SetExtended(extended)
			logging.Error(err)

		default:
			logging.Info(v.Name + " field in PICS ignored (Change " + strconv.Itoa(pack.PICSChangeID) + ")")
		}

		logging.Error(err)
	}

	// Update from API
	err = pack.Update()
	if err != nil && err != steam.ErrPackageNotFound {
		return false, true, err
	}

	// Save new data
	gorm.Save(&pack)
	if gorm.Error != nil {
		return false, true, gorm.Error
	}

	// Save price changes
	var prices db.ProductPrices
	var price db.ProductPriceCache
	var kinds []db.Kind
	for code := range steam.Countries {

		var oldPrice, newPrice int

		prices, err = packageBeforeUpdate.GetPrices()
		if err == nil {
			price, err = prices.Get(code)
			if err == nil {
				oldPrice = price.Final
			} else {
				continue // Only compare if there is an old price to compare to
			}
		}

		prices, err = pack.GetPrices()
		if err == nil {
			price, err = prices.Get(code)
			if err == nil {
				newPrice = price.Final
			}
		}

		if oldPrice != newPrice {
			kinds = append(kinds, db.CreateProductPrice(pack, code, oldPrice, newPrice))
		}
	}

	err = db.BulkSaveKinds(kinds, db.KindProductPrice, true)
	if err != nil {
		return false, true, err
	}

	// Send websocket
	page, err := websockets.GetPage(websockets.PagePackages)
	if err == nil && page.HasConnections() {

		page.Send(pack.OutputForJSON())
	}

	return true, false, nil
}
