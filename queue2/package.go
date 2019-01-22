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

func QueuePackage(IDs []int) (err error) {

	b, err := json.Marshal(producePackagePayload{
		ID:   IDs,
		Time: time.Now().Unix(),
	})
	if err != nil {
		return err
	}

	return Produce(QueuePackages, b)
}

// JSON must match the Updater app
type producePackagePayload struct {
	ID   []int `json:"IDs"`
	Time int64 `json:"Time"`
}

type RabbitMessagePackage struct {
	PICSPackageInfo RabbitMessageProduct
	Payload         producePackagePayload
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
	rabbitMessage := RabbitMessagePackage{}

	err = helpers.Unmarshal(msg.Body, &rabbitMessage)
	if err != nil {
		return false, err
	}

	message := rabbitMessage.PICSPackageInfo

	logInfo("Consuming package: " + strconv.Itoa(message.ID))

	if !db.IsValidPackageID(message.ID) {
		return false, errors.New("invalid package ID: " + strconv.Itoa(message.ID))
	}

	// Load current package
	gorm, err := db.GetMySQLClient()
	if err != nil {
		return true, err
	}

	pack := db.Package{}
	gorm.FirstOrInit(&pack, db.Package{ID: message.ID})
	if gorm.Error != nil {
		return true, gorm.Error
	}

	// Skip if updated in last day, unless its from PICS
	if pack.UpdatedAt.Unix() > time.Now().Add(time.Hour * -24).Unix() && pack.ChangeNumber >= message.ChangeNumber {
		logInfo("Skipping, updated in last day")
		return false, nil
	}

	var packageBeforeUpdate = pack

	// Update from PICS
	err = updatePackageFromPICS(&pack, rabbitMessage)
	if err != nil {
		return true, err
	}

	// Update from API
	err = updatePackageFromStore(&pack)
	err = helpers.IgnoreErrors(err, steam.ErrPackageNotFound)
	if err != nil {
		return true, err
	}

	// Set package name to app name
	if pack.AppsCount == 1 {

		appIDs, err := pack.GetAppIDs()
		if err != nil {
			return true, err
		}

		app, err := db.GetApp(appIDs[0], []string{})
		if err != nil && err != db.ErrRecordNotFound {
			return true, err
		} else if err == nil && pack.HasDefaultName() {
			pack.Name = app.Name
			pack.Icon = app.GetIcon()
		}
	}

	// Save price changes
	err = savePriceChanges(packageBeforeUpdate, pack)
	if err != nil {
		return true, err
	}

	// Save new data
	gorm = gorm.Save(&pack)
	if gorm.Error != nil {
		return true, gorm.Error
	}

	// Send websocket
	page, err := websockets.GetPage(websockets.PagePackage)
	if err != nil {
		return true, err
	} else if page.HasConnections() {
		page.Send(pack.ID)
	}

	return false, err
}

func updatePackageFromPICS(pack *db.Package, rabbitMessage RabbitMessagePackage) (err error) {

	message := rabbitMessage.PICSPackageInfo

	// Update with new details
	if message.ChangeNumber > pack.ChangeNumber {
		pack.ChangeNumberDate = time.Unix(rabbitMessage.Payload.Time, 0)
	}

	pack.ID = message.ID
	pack.ChangeNumber = message.ChangeNumber
	pack.Name = message.KeyValues.Name

	for _, v := range message.KeyValues.Children {

		switch v.Name {
		case "billingtype":

			var i64 int64
			i64, err = strconv.ParseInt(v.Value.(string), 10, 8)
			pack.BillingType = int8(i64)

		case "licensetype":

			var i64 int64
			i64, err = strconv.ParseInt(v.Value.(string), 10, 8)
			pack.LicenseType = int8(i64)

		case "status":

			var i64 int64
			i64, err = strconv.ParseInt(v.Value.(string), 10, 8)
			pack.Status = int8(i64)

		case "packageid":
			// Empty
		case "appids":

			apps := helpers.StringSliceToIntSlice(v.GetChildrenAsSlice())

			var b []byte
			b, err = json.Marshal(apps)

			if err == nil {
				pack.AppIDs = string(b)
				pack.AppsCount = len(apps)
			}

		case "depotids":

			depots := helpers.StringSliceToIntSlice(v.GetChildrenAsSlice())

			var b []byte
			b, err = json.Marshal(depots)

			if err == nil {
				pack.DepotIDs = string(b)
			}

		case "appitems":

			var appItems = map[string]string{}
			for _, vv := range v.Children {
				if len(vv.Children) == 1 {
					appItems[vv.Name] = vv.Children[0].Value.(string)
				}
			}

			var b []byte
			b, err = json.Marshal(appItems)

			if err == nil {
				pack.AppItems = string(b)
			}

		case "extended":

			var b []byte
			b, err = json.Marshal(v.GetExtended())

			if err == nil {
				pack.Extended = string(b)
			}

		default:
			err = errors.New(v.Name + " field in package PICS ignored (Change " + strconv.Itoa(pack.ChangeNumber) + ")")
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
		if err != nil && err != steam.ErrPackageNotFound {
			return err
		}

		prices.AddPriceFromPackage(code, response)

		if code == steam.CountryUS {

			// Controller
			b, err := json.Marshal(response.Data.Controller)
			if err != nil {
				return err
			}

			pack.Controller = string(b)

			// Platforms
			var platforms []string
			if response.Data.Platforms.Linux {
				platforms = append(platforms, "linux")
			}
			if response.Data.Platforms.Windows {
				platforms = append(platforms, "windows")
			}
			if response.Data.Platforms.Mac {
				platforms = append(platforms, "macos")
			}

			b, err = json.Marshal(platforms)
			if err != nil {
				return err
			}

			pack.Platforms = string(b)

			//
			pack.ImageHeader = response.Data.HeaderImage
			pack.ImageLogo = response.Data.SmallLogo
			pack.ImageHeader = response.Data.HeaderImage
			pack.ReleaseDate = response.Data.ReleaseDate.Date
			pack.ReleaseDateUnix = helpers.GetReleaseDateUnix(response.Data.ReleaseDate.Date)
			pack.ComingSoon = response.Data.ReleaseDate.ComingSoon
			pack.Name = response.Data.Name
		}
	}

	b, err := json.Marshal(prices)
	if err != nil {
		return err
	}

	pack.Prices = string(b)

	return nil
}
