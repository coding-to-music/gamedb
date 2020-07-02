package queue

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/go-durationfmt"
	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/Jleagle/steam-go/steamvdf"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql/pics"
	steamHelper "github.com/gamedb/gamedb/pkg/steam"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/gocolly/colly"
)

type PackageMessage struct {
	ID           int                    `json:"id"`
	ChangeNumber int                    `json:"change_number"`
	VDF          map[string]interface{} `json:"vdf"`
}

func packageHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := PackageMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		if payload.ID == 0 {
			message.Ack(false)
			continue
		}

		if !helpers.IsValidPackageID(payload.ID) {
			log.Err(err, payload.ID)
			sendToFailQueue(message)
			continue
		}

		// Load current package
		pack, err := mongo.GetPackage(payload.ID)
		if err == mongo.ErrNoDocuments {
			pack = mongo.Package{}
			pack.ID = payload.ID
		} else if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			continue
		}

		// Skip if updated in last day, unless its from PICS
		if !config.IsLocal() && !pack.ShouldUpdate() && pack.ChangeNumber >= payload.ChangeNumber {

			s, err := durationfmt.Format(time.Since(pack.UpdatedAt), "%hh %mm")
			log.Err(err)

			log.Info("Skipping package, updated " + s + " ago")
			message.Ack(false)
			continue
		}

		// Produce price changes
		if config.IsLocal() {

			for _, v := range i18n.GetProdCCs(true) {

				payload2 := PackagePriceMessage{
					PackageID:   uint(pack.ID),
					PackageName: pack.Name,
					PackageIcon: pack.Icon,
					ProductCC:   v.ProductCode,
					Time:        message.FirstSeen(),
					BeforePrice: nil,
				}

				price := pack.GetPrices().Get(v.ProductCode)
				if price.Exists {

					payload2.BeforePrice = &price.Final

					err = producePackagePrice(payload2)
					if err != nil {
						log.Err(err)
					}
				}
			}
		}

		//
		var packageBeforeUpdate = pack

		// Update from PICS
		err = updatePackageFromPICS(&pack, message, payload)
		if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			continue
		}

		var wg sync.WaitGroup

		// Update from store.steampowered.com JSON
		wg.Add(1)
		go func() {

			defer wg.Done()

			var err error

			err = updatePackageFromStore(&pack)
			err = helpers.IgnoreErrors(err, steamapi.ErrPackageNotFound)
			if err != nil {

				if err == steamapi.ErrHTMLResponse {
					log.Info(err, payload.ID)
				} else {
					steamHelper.LogSteamError(err, payload.ID)
				}

				sendToRetryQueue(message)
				return
			}
		}()

		// Scrape from store.steampowered.com
		wg.Add(1)
		go func() {

			defer wg.Done()

			var err = scrapePackage(&pack)
			if err != nil {
				steamHelper.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		// Set package name to app name
		wg.Add(1)
		go func() {

			defer wg.Done()

			var err = updatePackageNameFromApp(&pack)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		wg.Wait()

		if message.ActionTaken {
			continue
		}

		// Save price changes
		wg.Add(1)
		go func() {

			defer wg.Done()

			var err = saveProductPricesToMongo(packageBeforeUpdate, pack)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		// Save package
		wg.Add(1)
		go func() {

			defer wg.Done()

			var err = pack.Save()
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		wg.Wait()

		if message.ActionTaken {
			continue
		}

		// Send websocket
		wg.Add(1)
		go func() {

			defer wg.Done()

			if payload.ChangeNumber > 0 {

				var err error

				wsPayload := IntPayload{ID: payload.ID}
				err = ProduceWebsocket(wsPayload, websockets.PagePackage, websockets.PagePackages)
				if err != nil {
					log.Err(err, payload.ID)
				}
			}
		}()

		// Clear caches
		wg.Add(1)
		go func() {

			defer wg.Done()

			var items = []string{
				memcache.MemcachePackage(pack.ID).Key,
				memcache.MemcachePackageInQueue(pack.ID).Key,
				memcache.MemcachePackageBundles(pack.ID).Key,
			}

			err := memcache.Delete(items...)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		// Queue apps
		wg.Add(1)
		go func() {

			defer wg.Done()

			if payload.ChangeNumber > 0 {

				var err error

				for _, appID := range pack.Apps {
					err = ProducePackage(PackageMessage{ID: appID})
					err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
					log.Err(err)
				}
			}
		}()

		wg.Wait()

		if message.ActionTaken {
			continue
		}

		//
		message.Ack(false)
	}
}

func updatePackageNameFromApp(pack *mongo.Package) (err error) {

	if len(pack.Apps) == 1 {

		app, err := mongo.GetApp(pack.Apps[0])
		if err == nil && app.Name != "" && (pack.Name == "" || pack.Name == "Package "+strconv.Itoa(pack.ID) || pack.Name == strconv.Itoa(pack.ID)) {

			pack.SetName(app.GetName(), false)
			pack.Icon = app.GetIcon()

		} else if err == mongo.ErrNoDocuments {
			return nil
		} else {
			return err
		}
	}

	return nil
}

func updatePackageFromPICS(pack *mongo.Package, message *rabbit.Message, payload PackageMessage) (err error) {

	if payload.ChangeNumber == 0 || pack.ChangeNumber >= payload.ChangeNumber {
		return nil
	}

	var kv = steamvdf.FromMap(payload.VDF)

	pack.ID = payload.ID
	pack.ChangeNumber = payload.ChangeNumber
	pack.ChangeNumberDate = message.FirstSeen()

	// Reset values that might be removed
	pack.BillingType = 0
	pack.LicenseType = 0
	pack.Status = 0
	pack.Apps = []int{}
	pack.AppsCount = 0
	pack.Depots = []int{}
	pack.AppItems = map[int]int{}
	pack.Extended = pics.PICSKeyValues{}

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
			pack.BillingType = int(i64)

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

			pack.Apps = apps
			pack.AppsCount = len(apps)

		case "depotids":

			pack.Depots = helpers.StringSliceToIntSlice(child.GetChildrenAsSlice())

		case "appitems":

			var appItems = map[int]int{}

			if len(child.Children) > 1 {
				log.Warning("More app items", pack.ID)
			}

			for _, vv := range child.Children {
				if len(vv.Children) == 1 {

					i1, err := strconv.Atoi(vv.Key)
					if err == nil {
						i2, err := strconv.Atoi(vv.Children[0].Value)
						if err == nil {
							appItems[i1] = i2
						}
					}

					if len(vv.Children) > 1 {
						log.Warning("More app items2", pack.ID)
					}
				}
			}

			pack.AppItems = appItems

		case "extended":

			pack.Extended = child.GetChildrenAsMap()

		case "extendedz": // For package 439999

			if len(pack.Extended) == 0 {
				pack.Extended = child.GetChildrenAsMap()
			} else {
				log.Warning("extendedz", pack.ID, child)
			}

		case "extendedasdf": // For package 439981

			log.Info(child) // todo

		case "is_available": // For package 439981

			log.Info(child) // todo, bool

		case "":

			// Some packages (46028) have blank children

		default:
			log.Warning(child.Key + " field in package PICS ignored (Package: " + strconv.Itoa(pack.ID) + ")")
		}

		if err != nil {
			return err
		}
	}

	return nil
}

var packageRegex = regexp.MustCompile(`store\.steampowered\.com/sub/[0-9]+$`)

func scrapePackage(pack *mongo.Package) (err error) {

	pack.InStore = false

	c := colly.NewCollector(
		colly.URLFilters(packageRegex),
		colly.AllowURLRevisit(),
		steamHelper.WithTimeout,
	)

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

func updatePackageFromStore(pack *mongo.Package) (err error) {

	prices := helpers.ProductPrices{}

	for _, cc := range i18n.GetProdCCs(true) {

		// Get package details
		response, _, err := steamHelper.GetSteam().GetPackageDetails(uint(pack.ID), cc.ProductCode, steamapi.LanguageEnglish)
		err = steamHelper.AllowSteamCodes(err)
		if err == steamapi.ErrPackageNotFound {
			continue
		}
		if err != nil {
			return err
		}

		prices.AddPriceFromPackage(cc.ProductCode, response)

		if cc.ProductCode == steamapi.ProductCCUS {

			// Controller
			var controller = pics.PICSController{}
			for k, v := range response.Data.Controller {
				controller[k] = v
			}

			pack.Controller = controller

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

			pack.Platforms = platforms

			// Images
			var wg sync.WaitGroup

			wg.Add(1)
			go func() {

				defer wg.Done()

				code := helpers.HeadWithTimeout(response.Data.SmallLogo, 0)
				pack.ImageLogo = ""
				if code == 200 {
					pack.ImageLogo = response.Data.SmallLogo
				}
			}()

			wg.Add(1)
			go func() {

				defer wg.Done()

				code := helpers.HeadWithTimeout(response.Data.PageImage, 0)
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

	pack.Prices = prices

	return nil
}
