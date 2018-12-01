package queue

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
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

func (d RabbitMessageApp) process(msg amqp.Delivery) (requeue bool, err error) {

	// Get message payload
	rabbitMessage := new(RabbitMessageApp)

	err = helpers.Unmarshal(msg.Body, rabbitMessage)
	if err != nil {
		return false, err
	}

	message := rabbitMessage.PICSAppInfo

	queueLog(log.SeverityInfo, "Consuming app: "+strconv.Itoa(message.ID))

	if !db.IsValidAppID(message.ID) {
		return false, errors.New("invalid app ID: " + strconv.Itoa(message.ID))
	}

	// Load current app
	gorm, err := db.GetMySQLClient()
	if err != nil {
		return true, err
	}

	app := db.App{}
	gorm.First(&app, message.ID)
	if gorm.Error != nil && !gorm.RecordNotFound() {
		return true, gorm.Error
	}

	if app.PICSChangeNumber >= message.ChangeNumber {
		return false, nil
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
			queueLog(err)
			app.ID = int(i64)

		case "common":

			var common = db.PICSAppCommon{}
			for _, vv := range v.Children {
				if vv.Value == nil {
					bytes, err := json.Marshal(vv.ToNestedMaps())
					queueLog(err)
					common[vv.Name] = string(bytes)
				} else {
					common[vv.Name] = vv.Value.(string)
				}
			}
			err = app.SetCommon(common)
			queueLog(err)

		case "extended":

			err = app.SetExtended(v.GetExtended())
			queueLog(err)

		case "config":

			config, launch := v.GetAppConfig()

			err = app.SetConfig(config)
			queueLog(err)

			err = app.SetLaunch(launch)
			queueLog(err)

		case "depots":

			err = app.SetDepots(v.GetAppDepots())
			queueLog(err)

		case "public_only":

			if v.Value.(string) == "1" {
				app.PICSPublicOnly = true
			}

		case "ufs":

			var common = db.PICSAppUFS{}
			for _, vv := range v.Children {
				if vv.Value == nil {
					bytes, err := json.Marshal(vv.ToNestedMaps())
					queueLog(err)
					common[vv.Name] = string(bytes)
				} else {
					common[vv.Name] = vv.Value.(string)
				}
			}
			err = app.SetUFS(common)
			queueLog(err)

		case "install":

			err = app.SetInstall(v.ToNestedMaps())
			queueLog(err)

		case "localization":

			err = app.SetLocalization(v.ToNestedMaps())
			queueLog(err)

		case "sysreqs":

			err = app.SetSystemRequirements(v.ToNestedMaps())
			queueLog(err)

		default:
			queueLog(log.SeverityInfo, v.Name+" field in app PICS ignored (Change "+strconv.Itoa(app.PICSChangeNumber)+")")
		}
	}

	// Update from API
	errs := app.UpdateFromAPI()
	for _, v := range errs {
		queueLog(v)
	}
	for _, v := range errs {
		if v != nil && v != steam.ErrAppNotFound {
			return true, v
		}
	}

	// Save price changes
	err = savePriceChanges(appBeforeUpdate, app)
	if err != nil {
		return true, err
	}

	// Send websocket
	page, err := websockets.GetPage(websockets.PageApps)
	if err == nil && page.HasConnections() {
		page.Send(app.OutputForJSON(steam.CountryUS))
	}

	// Save new data
	gorm = gorm.Save(&app)
	if gorm.Error != nil {
		return true, gorm.Error
	}

	return false, err
}
