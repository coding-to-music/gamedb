package queue

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
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

func (d RabbitMessagePackage) process(msg amqp.Delivery) (requeue bool, err error) {

	// Get message
	rabbitMessage := new(RabbitMessagePackage)

	err = helpers.Unmarshal(msg.Body, rabbitMessage)
	if err != nil {
		return false, err
	}

	message := rabbitMessage.PICSPackageInfo

	queueLog("Consuming package: " + strconv.Itoa(message.ID))

	if !db.IsValidPackageID(message.ID) {
		return false, errors.New("invalid package ID: " + strconv.Itoa(message.ID))
	}

	// Load current package
	gorm, err := db.GetMySQLClient()
	if err != nil {
		return true, err
	}

	pack := db.Package{}
	gorm.First(&pack, message.ID)
	if gorm.Error != nil && !gorm.RecordNotFound() {
		return true, gorm.Error
	}

	var packageBeforeUpdate = pack

	// Update from PICS
	if pack.PICSChangeNumber < message.ChangeNumber {

		err = updatePackageFromPICS(&pack, message)
		if err != nil {
			return true, err
		}
	}

	// Update from API
	err = updatePackageFromStore(&pack)
	if err != nil && err != steam.ErrPackageNotFound {
		return true, err
	}

	// Set package name to app name
	if pack.AppsCount == 1 {

		apps, err := pack.GetAppIDs()
		if err != nil {
			return true, err
		}

		app, err := db.GetApp(apps[0])
		if err != nil && err != db.ErrCantFindApp {
			return true, err
		} else if err != db.ErrCantFindApp && pack.HasDefaultName() {
			pack.PICSName = app.Name
		}
	}

	// Save price changes
	err = savePriceChanges(packageBeforeUpdate, pack)
	if err != nil {
		return true, err
	}

	// Send websocket
	page, err := websockets.GetPage(websockets.PagePackages)
	if err == nil && page.HasConnections() {
		page.Send(pack.OutputForJSON(steam.CountryUS))
	}

	// Save new data
	gorm = gorm.Save(&pack)
	if gorm.Error != nil {
		return true, gorm.Error
	}

	return false, err
}

func updatePackageFromPICS(pack *db.Package, message RabbitMessageProduct) (err error) {

	// Update with new details
	if message.ChangeNumber > pack.PICSChangeNumber {
		pack.PICSChangeNumberDate = time.Now()
	}

	pack.ID = message.ID
	pack.PICSChangeNumber = message.ChangeNumber
	pack.PICSName = message.KeyValues.Name

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

			err = pack.SetAppIDs(helpers.StringSliceToIntSlice(v.GetChildrenAsSlice()))

		case "depotids":

			err = pack.SetDepotIDs(helpers.StringSliceToIntSlice(v.GetChildrenAsSlice()))

		case "appitems":

			var appItems = map[string]string{}
			for _, vv := range v.Children {
				if len(vv.Children) == 1 {
					appItems[vv.Name] = vv.Children[0].Value.(string)
				}
			}
			err = pack.SetAppItems(appItems)

		case "extended":

			err = pack.SetExtended(v.GetExtended())

		default:
			err = errors.New(v.Name + " field in package PICS ignored (Change " + strconv.Itoa(pack.PICSChangeNumber) + ")")
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func updatePackageFromStore(pack *db.Package) (err error) {

	prices := db.ProductPrices{}

	for _, code := range helpers.GetActiveCountries() {

		// Get package details
		response, _, err := helpers.GetSteam().GetPackageDetails(pack.ID, code, steam.LanguageEnglish)
		if err != nil {

			// Presume that if not found in one language, wont be found in any.
			if err == steam.ErrPackageNotFound {
				break
			}

			return err
		}

		prices.AddPriceFromPackage(code, response)

		if code == steam.CountryUS {

			// Controller
			controllerString, err := json.Marshal(response.Data.Controller)
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

			//
			pack.ImageHeader = response.Data.HeaderImage
			pack.ImageLogo = response.Data.SmallLogo
			pack.ImageHeader = response.Data.HeaderImage
			pack.Platforms = string(platformsString)
			pack.Controller = string(controllerString)
			pack.ReleaseDate = response.Data.ReleaseDate.Date
			pack.ReleaseDateUnix = helpers.GetReleaseDateUnix(response.Data.ReleaseDate.Date)
			pack.ComingSoon = response.Data.ReleaseDate.ComingSoon
		}
	}

	return pack.SetPrices(prices)
}
