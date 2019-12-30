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
	"github.com/Jleagle/valve-data-format-go/vdf"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	pubsubHelpers "github.com/gamedb/gamedb/pkg/helpers/pubsub"
	steamHelper "github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue/framework"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/sql/pics"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/gocolly/colly"
	influx "github.com/influxdata/influxdb1-client"
)

type PackageMessage struct {
	ID           int                    `json:"id"`
	ChangeNumber int                    `json:"change_number"`
	VDF          map[string]interface{} `json:"vdf"`
}

func packageHandler(messages []*framework.Message) {

	for _, message := range messages {

		payload := PackageMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			return
		}

		var id = payload.ID

		if !sql.IsValidPackageID(id) {
			log.Err(err, payload.ID)
			sendToFailQueue(message)
			return
		}

		// Load current package
		gorm, err := sql.GetMySQLClient()
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			return
		}

		pack := sql.Package{}
		gorm = gorm.FirstOrInit(&pack, sql.Package{ID: id})
		if gorm.Error != nil {
			log.Err(gorm.Error, payload.ID)
			sendToRetryQueue(message)
			return
		}

		var newPackage bool
		if pack.CreatedAt.IsZero() {
			newPackage = true
		}

		// Skip if updated in last day, unless its from PICS
		if !config.IsLocal() {
			if pack.UpdatedAt.After(time.Now().Add(time.Hour * 24 * -1)) {
				if pack.ChangeNumber >= payload.ChangeNumber {
					log.Info("Skipping package, updated in last day")
					message.Ack()
					return
				}
			}
		}

		var packageBeforeUpdate = pack

		// Update from PICS
		err = updatePackageFromPICS(&pack, message, payload)
		if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			return
		}

		// Update from API
		err = updatePackageFromStore(&pack)
		err = helpers.IgnoreErrors(err, steam.ErrPackageNotFound)
		if err != nil {

			if err == steam.ErrHTMLResponse {
				log.Info(err, id)
			} else {
				steamHelper.LogSteamError(err, id)
			}

			sendToRetryQueue(message)
			return
		}

		// Scrape
		err = scrapePackage(&pack)
		if err != nil {
			steamHelper.LogSteamError(err, payload.ID)
			sendToRetryQueue(message)
			return
		}

		// Set package name to app name
		err = updatePackageNameFromApp(&pack)
		if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			return
		}

		// Save price changes
		err = savePriceChanges(packageBeforeUpdate, pack)
		if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			return
		}

		// Save new data
		gorm = gorm.Save(&pack)
		if gorm.Error != nil {
			log.Err(gorm.Error, payload.ID)
			sendToRetryQueue(message)
			return
		}

		// Save to InfluxDB
		err = savePackageToInflux(pack)
		if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			return
		}

		// Send websocket
		wsPayload := websockets.PubSubIDPayload{}
		wsPayload.ID = pack.ID
		wsPayload.Pages = []websockets.WebsocketPage{websockets.PagePackage, websockets.PagePackages}

		_, err = pubsubHelpers.Publish(pubsubHelpers.PubSubTopicWebsockets, wsPayload)
		log.Err(err)

		// Clear caches
		var keys = []string{
			memcache.MemcachePackageInQueue(pack.ID).Key,
			memcache.MemcachePackageBundles(pack.ID).Key,
		}

		if pack.ReleaseDateUnix > time.Now().Unix() && newPackage {
			keys = append(keys, memcache.MemcacheUpcomingPackagesCount.Key)
		}

		err = memcache.RemoveKeyFromMemCacheViaPubSub(keys...)
		log.Err(err)

		// Queue apps
		// Commented out because queued too many apps
		// Uncommented out to help with finding sales
		for _, appID := range pack.GetAppIDs() {
			err = ProducePackage(PackageMessage{ID: appID})
			err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
			log.Err(err)
		}

		//
		message.Ack()
	}
}

func updatePackageNameFromApp(pack *sql.Package) (err error) {

	if pack.AppsCount == 1 {

		appIDs := pack.GetAppIDs()
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

func updatePackageFromPICS(pack *sql.Package, message *framework.Message, payload PackageMessage) (err error) {

	if payload.ChangeNumber == 0 || pack.ChangeNumber >= payload.ChangeNumber {
		return nil
	}

	var kv = vdf.FromMap(payload.VDF)

	pack.ID = payload.ID
	pack.ChangeNumber = payload.ChangeNumber
	pack.ChangeNumberDate = message.FirstSeen()

	// Reset values that might be removed
	pack.BillingType = 0
	pack.LicenseType = 0
	pack.Status = 0
	pack.AppIDs = "[]"
	pack.AppsCount = 0
	pack.DepotIDs = ""
	pack.AppItems = ""
	pack.Extended = ""

	if len(kv.Children) == 1 && kv.Children[0].Key == strconv.Itoa(payload.ID) {
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

			appIDs := helpers.StringSliceToIntSlice(child.GetChildrenAsSlice())

			// mongo.UpdatePackageApps()

			var b []byte
			b, err = json.Marshal(appIDs)

			if err == nil {
				pack.AppIDs = string(b)
				pack.AppsCount = len(appIDs)
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

		case "":

			// Some packages (46028) have blank children

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
		colly.AllowURLRevisit(),
	)
	c.SetRequestTimeout(time.Second * 60)

	// ID
	c.OnHTML("h2.pageheader", func(e *colly.HTMLElement) {
		pack.SetName(e.Text, true)
		pack.InStore = true
	})

	//
	c.OnError(func(r *colly.Response, err error) {
		steamHelper.LogSteamError(err)
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
		response, b, err := steamHelper.GetSteam().GetPackageDetails(pack.ID, cc.ProductCode, steam.LanguageEnglish)
		err = steamHelper.AllowSteamCodes(err, b, nil)
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

	_, err := influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementPackages),
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
