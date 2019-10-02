package queue

import (
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/Jleagle/valve-data-format-go"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/sql/pics"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/gocolly/colly"
	influx "github.com/influxdata/influxdb1-client"
	"github.com/streadway/amqp"
)

type packageMessage struct {
	baseMessage
	Message packageMessageInner `json:"message"`
}

type packageMessageInner struct {
	ID           int                    `json:"id"`
	ChangeNumber int                    `json:"change_number,omitempty"`
	VDF          map[string]interface{} `json:"vdf,omitempty"`
}

type packageQueue struct {
}

func (q packageQueue) processMessages(msgs []amqp.Delivery) {

	msg := msgs[0]

	var err error

	message := packageMessage{}
	message.OriginalQueue = queuePackages

	err = helpers.Unmarshal(msg.Body, &message)
	if err != nil {
		log.Err(err, msg.Body)
		ackFail(msg, &message)
		return
	}

	var id = message.Message.ID

	if message.Attempt > 1 {
		log.Info("Consuming package " + strconv.Itoa(id) + ", attempt " + strconv.Itoa(message.Attempt))
	}

	if !sql.IsValidPackageID(id) {
		log.Info(errors.New("invalid package ID: "+strconv.Itoa(id)), msg.Body)
		ackFail(msg, &message)
		return
	}

	// Load current package
	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err, id)
		ackRetry(msg, &message)
		return
	}

	pack := sql.Package{}
	gorm = gorm.FirstOrInit(&pack, sql.Package{ID: id})
	if gorm.Error != nil {
		log.Err(gorm.Error, id)
		ackRetry(msg, &message)
		return
	}

	var newPackage bool
	if pack.CreatedAt.IsZero() {
		newPackage = true
	}

	// Skip if updated in last day, unless its from PICS
	if !message.Force {
		if !config.IsLocal() {
			if pack.UpdatedAt.Unix() > time.Now().Add(time.Hour*24*-1).Unix() {
				if pack.ChangeNumber >= message.Message.ChangeNumber && message.Message.ChangeNumber > 0 {
					log.Info("Skipping package, updated in last day")
					message.ack(msg)
					return
				}
			}
		}
	}

	var packageBeforeUpdate = pack

	// Update from PICS
	err = updatePackageFromPICS(&pack, message)
	if err != nil {
		log.Err(err, id)
		ackRetry(msg, &message)
		return
	}

	// Update from API
	err = updatePackageFromStore(&pack)
	err = helpers.IgnoreErrors(err, steam.ErrPackageNotFound)
	if err != nil {

		if err == steam.ErrHTMLResponse {
			log.Info(err, id)
		} else {
			helpers.LogSteamError(err, id)
		}

		ackRetry(msg, &message)
		return
	}

	// Scrape
	err = scrapePackage(&pack)
	if err != nil {
		log.Err(err, id)
		ackRetry(msg, &message)
		return
	}

	// Set package name to app name
	err = updatePackageNameFromApp(&pack)
	if err != nil {
		log.Err(err, id)
		ackRetry(msg, &message)
		return
	}

	// Save price changes
	err = savePriceChanges(packageBeforeUpdate, pack)
	if err != nil {
		log.Err(err, id)
		ackRetry(msg, &message)
		return
	}

	// Save new data
	gorm = gorm.Save(&pack)
	if gorm.Error != nil {
		log.Err(gorm.Error, id)
		ackRetry(msg, &message)
		return
	}

	// Save to InfluxDB
	err = savePackageToInflux(pack)
	if err != nil {
		log.Err(err, id)
		ackRetry(msg, &message)
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
			helpers.MemcachePackageBundles(pack.ID).Key,
		)
		log.Err(err)
	}

	message.ack(msg)
}

func updatePackageNameFromApp(pack *sql.Package) (err error) {

	if pack.AppsCount == 1 {

		appIDs, err := pack.GetAppIDs()
		if err != nil {
			return err
		}

		app, err := sql.GetApp(appIDs[0], nil)
		if err == nil && app.Name != "" && (pack.Name == "" || pack.Name == "Package "+strconv.Itoa(pack.ID) || pack.Name == strconv.Itoa(pack.ID)) {

			pack.SetName(app.GetName(), false)
			pack.Icon = app.GetIcon()

		} else if err == sql.ErrRecordNotFound {
			return nil
		} else {
			return err
		}
	}

	return nil
}

func updatePackageFromPICS(pack *sql.Package, message packageMessage) (err error) {

	if message.Message.ChangeNumber == 0 {
		return nil
	}

	if pack.ChangeNumber >= message.Message.ChangeNumber {
		return nil
	}

	var kv = vdf.FromMap(message.Message.VDF)

	pack.ID = message.Message.ID
	pack.ChangeNumber = message.Message.ChangeNumber
	pack.ChangeNumberDate = message.FirstSeen

	// Reset values that might be removed
	pack.BillingType = 0
	pack.LicenseType = 0
	pack.Status = 0
	pack.AppIDs = ""
	pack.AppsCount = 0
	pack.DepotIDs = ""
	pack.AppItems = ""
	pack.Extended = ""

	if len(kv.Children) == 1 && kv.Children[0].Key == strconv.Itoa(message.Message.ID) {
		kv = kv.Children[0]
	}

	if len(kv.Children) == 0 {
		return nil
	}

	for _, child := range kv.Children {

		switch child.Key {
		case "billingtype":

			var i64 int64
			i64, err = strconv.ParseInt(child.Value, 10, 8)
			pack.BillingType = int8(i64)

		case "licensetype":

			var i64 int64
			i64, err = strconv.ParseInt(child.Value, 10, 8)
			pack.LicenseType = int8(i64)

		case "status":

			var i64 int64
			i64, err = strconv.ParseInt(child.Value, 10, 8)
			pack.Status = int8(i64)

		case "packageid":
			// Empty
		case "appids":

			apps := helpers.StringSliceToIntSlice(child.GetChildrenAsSlice())

			var b []byte
			b, err = json.Marshal(apps)

			if err == nil {
				pack.AppIDs = string(b)
				pack.AppsCount = len(apps)
			}

		case "depotids":

			depots := helpers.StringSliceToIntSlice(child.GetChildrenAsSlice())

			var b []byte
			b, err = json.Marshal(depots)

			if err == nil {
				pack.DepotIDs = string(b)
			}

		case "appitems":

			var appItems = map[string]string{}
			for _, vv := range child.Children {
				if len(vv.Children) == 1 {
					appItems[vv.Key] = vv.Children[0].Value
				}
			}

			var b []byte
			b, err = json.Marshal(appItems)

			if err == nil {
				pack.AppItems = string(b)
			}

		case "extended":

			var b []byte
			b, err = json.Marshal(child.GetChildrenAsMap())

			if err == nil {
				pack.Extended = string(b)
			}

		default:
			err = errors.New(child.Key + " field in package PICS ignored (Package: " + strconv.Itoa(pack.ID) + ")")
		}

		if err != nil {
			return err
		}
	}

	return nil
}

var packageRegex = regexp.MustCompile(`store\.steampowered\.com/sub/[0-9]+$`)

func scrapePackage(pack *sql.Package) (err error) {

	pack.InStore = false

	c := colly.NewCollector(
		colly.URLFilters(packageRegex),
	)

	// ID64
	c.OnHTML("h2.pageheader", func(e *colly.HTMLElement) {
		pack.SetName(e.Text, true)
		pack.InStore = true
	})

	//
	c.OnError(func(r *colly.Response, err error) {
		helpers.LogSteamError(err)
	})

	err = c.Visit("https://store.steampowered.com/sub/" + strconv.Itoa(pack.ID))
	if err != nil && strings.Contains(err.Error(), "because its not in AllowedDomains") {
		log.Info(err)
		return nil
	}

	return err
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
			pack.SetName(response.Data.Name, false)
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
