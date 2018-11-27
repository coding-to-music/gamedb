package queue

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

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

	if message.ChangeNumber > app.PICSChangeNumber {
		app.PICSChangeNumberDate = time.Now()
	}

	app.PICSChangeNumber = message.ChangeNumber
	app.Name = message.KeyValues.Name

	for _, v := range message.KeyValues.Children {

		switch v.Name {
		case "appid":

			var i64 int64
			i64, err = strconv.ParseInt(v.Value.(string), 10, 32)
			logging.Error(err)
			app.ID = int(i64)

		case "common":

			var common = db.PICSAppCommon{}
			for _, vv := range v.Children {
				if vv.Value == nil {
					bytes, err := json.Marshal(vv.ToNestedMaps())
					logging.Error(err)
					common[vv.Name] = string(bytes)
				} else {
					common[vv.Name] = vv.Value.(string)
				}
			}
			err = app.SetCommon(common)
			logging.Error(err)

		case "extended":

			err = app.SetExtended(v.GetExtended())
			logging.Error(err)

		case "config":

			config, launch := v.GetAppConfig()

			err = app.SetConfig(config)
			logging.Error(err)

			err = app.SetLaunch(launch)
			logging.Error(err)

		case "depots":

			err = app.SetDepots(v.GetAppDepots())
			logging.Error(err)

		case "public_only":

			if v.Value.(string) == "1" {
				app.PICSPublicOnly = true
			}

		case "ufs":

			var common = db.PICSAppUFS{}
			for _, vv := range v.Children {
				if vv.Value == nil {
					bytes, err := json.Marshal(vv.ToNestedMaps())
					logging.Error(err)
					common[vv.Name] = string(bytes)
				} else {
					common[vv.Name] = vv.Value.(string)
				}
			}
			err = app.SetUFS(common)
			logging.Error(err)

		case "install":

			err = app.SetInstall(v.ToNestedMaps())
			logging.Error(err)

		case "localization":

			err = app.SetLocalization(v.ToNestedMaps())
			logging.Error(err)

		default:
			logging.Info(v.Name + " field in PICS ignored (Change " + strconv.Itoa(app.PICSChangeNumber) + ")")
		}
	}

	// Update from API
	errs := app.UpdateFromAPI()
	for _, v := range errs {
		logging.Error(v)
	}
	for _, v := range errs {
		if v != nil && v != steam.ErrAppNotFound {
			return false, true, v
		}
	}

	// Save new data
	gorm = gorm.Save(&app)
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
			} else {
				continue // Only compare if there is a new price to compare to
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
		page.Send(app.OutputForJSON(steam.CountryUS))
	}

	return true, false, err
}
