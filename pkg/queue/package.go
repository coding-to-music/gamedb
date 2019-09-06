package queue

import (
	"encoding/json"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/sql/pics"
	"github.com/gamedb/gamedb/pkg/websockets"
	influx "github.com/influxdata/influxdb1-client"
	"github.com/mitchellh/mapstructure"
	"github.com/streadway/amqp"
)

type packageMessage struct {
	ID           int                    `json:"id"`
	ChangeNumber int                    `json:"change_number,omitempty"`
	VDF          map[string]interface{} `json:"vdf,omitempty"`
}

type packageQueue struct {
}

func (q packageQueue) processMessages(msgs []amqp.Delivery) {

	msg := msgs[0]

	var err error
	var payload = baseMessage{
		Message:       packageMessage{},
		OriginalQueue: queueGoPackages,
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

	if !sql.IsValidPackageID(message.ID) {
		logError(errors.New("invalid package ID: " + strconv.Itoa(message.ID)))
		payload.ack(msg)
		return
	}

	// Load current package
	gorm, err := sql.GetMySQLClient()
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	pack := sql.Package{}
	gorm = gorm.FirstOrInit(&pack, sql.Package{ID: message.ID})
	if gorm.Error != nil {
		logError(gorm.Error, message.ID)
		payload.ackRetry(msg)
		return
	}

	var newPackage bool
	if pack.CreatedAt.IsZero() {
		newPackage = true
	}

	// Skip if updated in last day, unless its from PICS
	if !config.IsLocal() {
		if pack.UpdatedAt.Unix() > time.Now().Add(time.Hour * 24 * -1).Unix() {
			if pack.ChangeNumber >= message.ChangeNumber && message.ChangeNumber > 0 {
				logInfo("Skipping package, updated in last day")
				payload.ack(msg)
				return
			}
		}
	}

	var packageBeforeUpdate = pack

	// Update from PICS
	err = updatePackageFromPICS(&pack, payload, message)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	// Update from API
	err = updatePackageFromStore(&pack)
	err = helpers.IgnoreErrors(err, steam.ErrPackageNotFound)
	if err != nil {

		if err == steam.ErrHTMLResponse {
			logInfo(err, message.ID)
		} else {
			helpers.LogSteamError(err, message.ID)
		}

		payload.ackRetry(msg)
		return
	}

	// Set package name to app name
	err = updatePackageNameFromApp(&pack)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	// Save price changes
	err = savePriceChanges(packageBeforeUpdate, pack)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	// Save new data
	gorm = gorm.Save(&pack)
	if gorm.Error != nil {
		logError(gorm.Error, message.ID)
		payload.ackRetry(msg)
		return
	}

	// Save to InfluxDB
	err = savePackageToInflux(pack)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	// Send websocket
	wsPayload := websockets.PubSubIDPayload{}
	wsPayload.ID = pack.ID
	wsPayload.Pages = []websockets.WebsocketPage{websockets.PagePackage, websockets.PagePackages}

	_, err = helpers.Publish(helpers.PubSubTopicWebsockets, wsPayload)
	log.Err(err)

	// Clear caches
	if pack.ReleaseDateUnix > time.Now().Unix() && newPackage {

		err = helpers.RemoveKeyFromMemCacheViaPubSub(
			helpers.MemcacheUpcomingPackagesCount.Key,
			helpers.MemcachePackageInQueue(pack.ID).Key,
		)
		log.Err(err)
	}

	payload.ack(msg)
}

func updatePackageNameFromApp(pack *sql.Package) (err error) {

	if pack.AppsCount == 1 {

		appIDs, err := pack.GetAppIDs()
		if err != nil {
			return err
		}

		app, err := sql.GetApp(appIDs[0], nil)
		if err == nil && app.Name != "" && (pack.Name == "" || pack.Name == "Package "+strconv.Itoa(pack.ID) || pack.Name == strconv.Itoa(pack.ID)) {

			pack.Name = app.GetName()
			pack.Icon = app.GetIcon()

		} else if err == sql.ErrRecordNotFound {
			return nil
		} else {
			return err
		}
	}

	return nil
}

func updatePackageFromPICS(pack *sql.Package, payload baseMessage, message packageMessage) (err error) {

	var vdf = parseVDF("root", message.VDF)

	if pack.ChangeNumber > message.ChangeNumber {
		return nil
	}

	pack.ID = message.ID
	// pack.Name = message.PICSPackageInfo.KeyValues.Name // todo
	pack.ChangeNumber = message.ChangeNumber
	pack.ChangeNumberDate = payload.FirstSeen

	// Reset values that might be removed
	pack.BillingType = 0
	pack.LicenseType = 0
	pack.Status = 0
	pack.AppIDs = ""
	pack.AppsCount = 0
	pack.DepotIDs = ""
	pack.AppItems = ""
	pack.Extended = ""

	for _, v := range vdf.Children {

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

func updatePackageFromStore(pack *sql.Package) (err error) {

	prices := sql.ProductPrices{}

	for _, cc := range helpers.GetProdCCs(true) {

		// Get package details
		response, b, err := helpers.GetSteam().GetPackageDetails(pack.ID, cc.ProductCode, steam.LanguageEnglish)
		err = helpers.AllowSteamCodes(err, b, nil)
		if err == steam.ErrPackageNotFound {
			continue
		}
		if err != nil {
			return err
		}

		prices.AddPriceFromPackage(cc.ProductCode, response)

		if cc.ProductCode == steam.ProductCCUS {

			// Controller
			var controller = pics.PICSController{}
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

				defer wg.Done()

				code := helpers.GetResponseCode(response.Data.SmallLogo)
				pack.ImageLogo = ""
				if code == 200 {
					pack.ImageLogo = response.Data.SmallLogo
				}
			}()

			wg.Add(1)
			go func() {

				defer wg.Done()

				code := helpers.GetResponseCode(response.Data.PageImage)
				pack.ImagePage = ""
				if code == 200 {
					pack.ImagePage = response.Data.PageImage
				}
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

func savePackageToInflux(pack sql.Package) error {

	price := pack.GetPrice(steam.ProductCCUS)
	if !price.Exists {
		return nil
	}

	_, err := helpers.InfluxWrite(helpers.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(helpers.InfluxMeasurementPackages),
		Tags: map[string]string{
			"package_id": strconv.Itoa(pack.ID),
		},
		Fields: map[string]interface{}{
			"price_us_initial":    price.Initial,
			"price_us_final":      price.Final,
			"price_us_discount":   price.DiscountPercent,
			"price_us_individual": price.Individual,
		},
		Time:      time.Now(),
		Precision: "m",
	})

	return err
}
