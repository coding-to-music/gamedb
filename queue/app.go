package queue

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
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
		log.Log(log.SeverityInfo, "Skipping app (Change number already processed)")
		return false, nil
	}

	var appBeforeUpdate = app

	err = updateAppPICS(&app, message)
	if err != nil {
		return true, err
	}

	err = updateAppDetails(&app)
	if err != nil && err != steam.ErrAppNotFound {
		return true, err
	}

	err = updateAppAchievements(&app)
	if err != nil {
		return true, err
	}

	err = updateAppSchema(&app)
	if err != nil {
		return true, err
	}

	errs := app.UpdateFromRequest("")
	for _, v := range errs {
		log.Log(v) // todo, requeue here if no errors should return from UpdateFromRequest
	}

	// Save price changes
	err = savePriceChanges(appBeforeUpdate, app)
	if err != nil {
		return true, err
	}

	// Send websocket
	page, err := websockets.GetPage(websockets.PageApps)
	if err != nil {
		return true, err
	} else if page.HasConnections() {
		page.Send(app.OutputForJSON(steam.CountryUS)) // todo, send one record with an array of all prices
	}

	// Misc
	app.Type = strings.ToLower(app.Type)
	app.ReleaseState = strings.ToLower(app.ReleaseState)

	// Save new data
	gorm = gorm.Save(&app)
	if gorm.Error != nil {
		return true, gorm.Error
	}

	return false, nil
}

func updateAppPICS(app *db.App, message RabbitMessageProduct) (err error) {

	if message.ChangeNumber > app.PICSChangeNumber {
		app.PICSChangeNumberDate = time.Now()
	}

	app.ID = message.ID
	app.PICSChangeNumber = message.ChangeNumber
	app.Name = message.KeyValues.Name

	for _, v := range message.KeyValues.Children {

		switch v.Name {
		case "appid":

			var i64 int64
			i64, err = strconv.ParseInt(v.Value.(string), 10, 32)
			if err != nil {
				return err
			}
			app.ID = int(i64)

		case "common":

			var common = db.PICSAppCommon{}
			for _, vv := range v.Children {
				if vv.Value == nil {
					bytes, err := json.Marshal(vv.ToNestedMaps())
					if err != nil {
						return err
					}
					common[vv.Name] = string(bytes)
				} else {
					common[vv.Name] = vv.Value.(string)
				}
			}
			err = app.SetCommon(common)
			if err != nil {
				return err
			}

		case "extended":

			err = app.SetExtended(v.GetExtended())
			if err != nil {
				return err
			}

		case "config":

			config, launch := v.GetAppConfig()

			err = app.SetConfig(config)
			if err != nil {
				return err
			}

			err = app.SetLaunch(launch)
			if err != nil {
				return err
			}

		case "depots":

			err = app.SetDepots(v.GetAppDepots())
			if err != nil {
				return err
			}

		case "public_only":

			if v.Value.(string) == "1" {
				app.PICSPublicOnly = true
			}

		case "ufs":

			var common = db.PICSAppUFS{}
			for _, vv := range v.Children {
				if vv.Value == nil {
					bytes, err := json.Marshal(vv.ToNestedMaps())
					if err != nil {
						return err
					}
					common[vv.Name] = string(bytes)
				} else {
					common[vv.Name] = vv.Value.(string)
				}
			}
			err = app.SetUFS(common)
			if err != nil {
				return err
			}

		case "install":

			err = app.SetInstall(v.ToNestedMaps())
			if err != nil {
				return err
			}

		case "localization":

			err = app.SetLocalization(v.ToNestedMaps())
			if err != nil {
				return err
			}

		case "sysreqs":

			err = app.SetSystemRequirements(v.ToNestedMaps())
			if err != nil {
				return err
			}

		default:
			queueLog(log.SeverityInfo, v.Name+" field in app PICS ignored (Change "+strconv.Itoa(app.PICSChangeNumber)+")")
		}
	}

	return nil
}

func updateAppDetails(app *db.App) error {

	prices := db.ProductPrices{}

	for _, code := range helpers.GetActiveCountries() {

		// Get app details
		response, _, err := helpers.GetSteam().GetAppDetails(app.ID, code, steam.LanguageEnglish)
		if err != nil {

			// Presume that if not found in one language, wont be found in any.
			if err == steam.ErrAppNotFound {
				break
			}

			return err
		}

		prices.AddPriceFromApp(code, response)

		if code == steam.CountryUS {

			// Screenshots
			screenshotsString, err := json.Marshal(response.Data.Screenshots)
			if err != nil {
				return err
			}

			// Movies
			moviesString, err := json.Marshal(response.Data.Movies)
			if err != nil {
				return err
			}

			// Achievements
			achievementsString, err := json.Marshal(response.Data.Achievements)
			if err != nil {
				return err
			}

			// DLC
			dlcString, err := json.Marshal(response.Data.DLC)
			if err != nil {
				return err
			}

			// Packages
			packagesString, err := json.Marshal(response.Data.Packages)
			if err != nil {
				return err
			}

			// Publishers
			publishersString, err := json.Marshal(response.Data.Publishers)
			if err != nil {
				return err
			}

			// Developers
			developersString, err := json.Marshal(response.Data.Developers)
			if err != nil {
				return err
			}

			// Categories
			var categories []int8
			for _, v := range response.Data.Categories {
				categories = append(categories, v.ID)
			}

			categoriesString, err := json.Marshal(categories)
			if err != nil {
				return err
			}

			genresString, err := json.Marshal(response.Data.Genres)
			if err != nil {
				return err
			}

			// Platforms
			var platforms []string
			if response.Data.Platforms.Linux {
				platforms = append(platforms, "linux")
			}
			if response.Data.Platforms.Windows {
				platforms = append(platforms, "windows")
			}
			if response.Data.Platforms.Windows {
				platforms = append(platforms, "macos")
			}

			platformsString, err := json.Marshal(platforms)
			if err != nil {
				return err
			}

			// Other
			app.Name = response.Data.Name
			app.Type = response.Data.Type
			app.IsFree = response.Data.IsFree
			app.DLC = string(dlcString)
			app.DLCCount = len(response.Data.DLC)
			app.ShortDescription = response.Data.ShortDescription
			app.HeaderImage = response.Data.HeaderImage
			app.Developers = string(developersString)
			app.Publishers = string(publishersString)
			app.Packages = string(packagesString)
			app.MetacriticScore = response.Data.Metacritic.Score
			app.MetacriticURL = response.Data.Metacritic.URL
			app.Categories = string(categoriesString)
			app.Genres = string(genresString)
			app.Screenshots = string(screenshotsString)
			app.Movies = string(moviesString)
			app.Achievements = string(achievementsString)
			app.Background = response.Data.Background
			app.Platforms = string(platformsString)
			app.GameID = response.Data.Fullgame.AppID
			app.GameName = response.Data.Fullgame.Name
			app.ReleaseDate = response.Data.ReleaseDate.Date
			app.ReleaseDateUnix = app.GetReleaseDateUnix() // Must be after setting app.ReleaseDate
			app.ComingSoon = response.Data.ReleaseDate.ComingSoon
		}
	}

	return app.SetPrices(prices)
}

func updateAppAchievements(app *db.App) error {

	percentages, _, err := helpers.GetSteam().GetGlobalAchievementPercentagesForApp(app.ID)

	// This endpoint seems to error if the app has no achievement data, so it's probably fine.
	err2, ok := err.(steam.Error)
	if ok && (err2.Code() == 403 || err2.Code() == 500) {
		return nil
	}

	percentagesString, err := json.Marshal(percentages)
	if err != nil {
		return err
	}

	app.AchievementPercentages = string(percentagesString)

	return nil
}

func updateAppSchema(app *db.App) error {

	schema, _, err := helpers.GetSteam().GetSchemaForGame(app.ID)

	// This endpoint seems to error if the app has no schema, so it's probably fine.
	err2, ok := err.(steam.Error)
	if ok && (err2.Code() == 403) {
		return nil
	}

	schemaString, err := json.Marshal(schema)
	if err != nil {
		return err
	}

	app.Schema = string(schemaString)

	return nil
}
