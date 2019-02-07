package queue

import (
	"encoding/json"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/websockets"
	"github.com/mitchellh/mapstructure"
	"github.com/streadway/amqp"
)

type packageMessage struct {
	ID              int                  `json:"id"`
	PICSPackageInfo rabbitMessageProduct `json:"pics_package_info"`
}

type packageQueue struct {
	baseQueue
}

func (q packageQueue) processMessage(msg amqp.Delivery) {

	var err error
	var payload = baseMessage{
		Message: packageMessage{},
	}

	err = helpers.Unmarshal(msg.Body, &payload)
	if err != nil {
		logError(err)
		payload.ack(msg)
		return
	}

	var message packageMessage
	err = mapstructure.Decode(payload.Message, &message)
	if err != nil {
		logError(err)
		payload.ack(msg)
		return
	}

	if payload.Attempt > 1 {
		logInfo("Consuming package " + strconv.Itoa(message.ID) + ", attempt " + strconv.Itoa(payload.Attempt))
	}

	if !db.IsValidPackageID(message.ID) {
		logError(errors.New("invalid package ID: " + strconv.Itoa(message.ID)))
		payload.ack(msg)
		return
	}

	// Load current package
	gorm, err := db.GetMySQLClient()
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	pack := db.Package{}
	gorm = gorm.FirstOrInit(&pack, db.Package{ID: message.ID})
	if gorm.Error != nil {
		logError(gorm.Error)
		payload.ackRetry(msg)
		return
	}

	// Skip if updated in last day, unless its from PICS
	if pack.UpdatedAt.Unix() > time.Now().Add(time.Hour * -24).Unix() && pack.ChangeNumber >= message.PICSPackageInfo.ChangeNumber {

		logInfo("Skipping package, updated in last day")
		if !config.Config.IsLocal() {
			payload.ack(msg)
			return
		}
	}

	var packageBeforeUpdate = pack

	// Update from PICS
	err = updatePackageFromPICS(&pack, payload, message)
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	// Update from API
	err = updatePackageFromStore(&pack)
	err = helpers.IgnoreErrors(err, steam.ErrPackageNotFound)
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	// Set package name to app name
	if pack.AppsCount == 1 {

		appIDs, err := pack.GetAppIDs()
		if err != nil {
			logError(err)
			payload.ackRetry(msg)
			return
		}

		app, err := db.GetApp(appIDs[0], []string{})
		if err != nil && err != db.ErrRecordNotFound {
			logError(err)
			payload.ackRetry(msg)
			return
		} else if err == nil && pack.HasDefaultName() {
			pack.Name = app.Name
			pack.Icon = app.GetIcon()
		}
	}

	// Save price changes
	err = savePriceChanges(packageBeforeUpdate, pack)
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	// Save new data
	gorm = gorm.Save(&pack)
	if gorm.Error != nil {
		logError(gorm.Error)
		payload.ackRetry(msg)
		return
	}

	// Send websocket
	page, err := websockets.GetPage(websockets.PagePackage)
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	if page.HasConnections() {
		page.Send(pack.ID)
	}

	payload.ack(msg)
}

func updatePackageFromPICS(pack *db.Package, payload baseMessage, message packageMessage) (err error) {

	if pack.ChangeNumber > message.PICSPackageInfo.ChangeNumber {
		return nil
	}

	pack.ID = message.ID
	pack.Name = message.PICSPackageInfo.KeyValues.Name
	pack.ChangeNumber = message.PICSPackageInfo.ChangeNumber
	pack.ChangeNumberDate = payload.FirstSeen

	for _, v := range message.PICSPackageInfo.KeyValues.Children {

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
			var controller = db.PICSController{}
			for k, v := range response.Data.Controller {
				controller[k] = v
			}

			b, err := json.Marshal(controller)
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

			// Images
			var wg sync.WaitGroup

			wg.Add(1)
			go func() {
				code := helpers.GetResponseCode(response.Data.HeaderImage)
				pack.ImageHeader = ""
				if code == 200 {
					pack.ImageHeader = response.Data.HeaderImage
				}
				wg.Done()
			}()

			wg.Add(1)
			go func() {
				code := helpers.GetResponseCode(response.Data.SmallLogo)
				pack.ImageLogo = ""
				if code == 200 {
					pack.ImageLogo = response.Data.SmallLogo
				}
				wg.Done()
			}()

			wg.Add(1)
			go func() {
				code := helpers.GetResponseCode(response.Data.PageImage)
				pack.ImagePage = ""
				if code == 200 {
					pack.ImagePage = response.Data.PageImage
				}
				wg.Done()
			}()

			wg.Wait()

			//
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
