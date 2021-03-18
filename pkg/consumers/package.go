package consumers

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

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
	"github.com/gamedb/gamedb/pkg/steam"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/gocolly/colly/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type PackageMessage struct {
	ID           int                    `json:"id"`
	ChangeNumber int                    `json:"change_number"`
	VDF          map[string]interface{} `json:"vdf"`
}

func packageHandler(message *rabbit.Message) {

	payload := PackageMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	if !helpers.IsValidPackageID(payload.ID) {
		message.Ack()
		return
	}

	// Load current package
	pack, err := mongo.GetPackage(payload.ID)
	if err == mongo.ErrNoDocuments {
		pack = mongo.Package{}
		pack.ID = payload.ID
	} else if err != nil {
		log.ErrS(err, payload.ID)
		sendToRetryQueue(message)
		return
	}

	// Skip if updated in last day, unless its from PICS
	if !config.IsLocal() && !pack.ShouldUpdate() && pack.ChangeNumber >= payload.ChangeNumber {

		// s, err := durationfmt.Format(time.Since(pack.UpdatedAt), "%hh %mm")
		// if err != nil {
		// 	log.ErrS(err)
		// }
		//
		// log.InfoS("Skipping package, updated " + s + " ago")

		message.Ack()
		return
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
					log.ErrS(err)
				}
			}
		}
	}

	//
	var packageBeforeUpdate = pack

	// Update from PICS
	err = updatePackageFromPICS(&pack, message, payload)
	if err != nil {
		log.ErrS(err, payload.ID)
		sendToRetryQueue(message)
		return
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
				log.InfoS(err, payload.ID)
			} else {
				steam.LogSteamError(err, zap.Int("package id", payload.ID))
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
			steam.LogSteamError(err, zap.Int("package id", payload.ID))
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
			log.ErrS(err, payload.ID)
			sendToRetryQueue(message)
			return
		}
	}()

	wg.Wait()

	if message.ActionTaken {
		return
	}

	// Save price changes
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err = saveProductPricesToMongo(packageBeforeUpdate, pack)
		if err != nil {
			log.ErrS(err, payload.ID)
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
			log.ErrS(err, payload.ID)
			sendToRetryQueue(message)
			return
		}
	}()

	wg.Wait()

	if message.ActionTaken {
		return
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
				log.ErrS(err, payload.ID)
			}
		}
	}()

	// Clear caches
	wg.Add(1)
	go func() {

		defer wg.Done()

		var items = []string{
			memcache.ItemPackage(pack.ID).Key,
			memcache.ItemPackageInQueue(pack.ID).Key,
			memcache.ItemPackageBundles(pack.ID).Key,
		}

		err := memcache.Client().Delete(items...)
		if err != nil {
			log.ErrS(err, payload.ID)
			sendToRetryQueue(message)
			return
		}
	}()

	// Queue apps
	// wg.Add(1)
	// go func() {
	//
	//	defer wg.Done()
	//
	//	if payload.ChangeNumber > 0 {
	//
	//		err := ProduceSteam(SteamMessage{AppIDs: pack.Apps})
	//		if err != nil {
	//			log.ErrS(err)
	//		}
	//	}
	// }()

	wg.Wait()

	if message.ActionTaken {
		return
	}

	//
	message.Ack()
}
func updatePackageNameFromApp(pack *mongo.Package) (err error) {

	if pack.HasEmptyName() || pack.HasEmptyIcon() || pack.ImageLogo == "" || pack.ImagePage == "" {

		apps, err := mongo.GetAppsByID(pack.Apps, bson.M{"_id": 1, "player_peak_alltime": 1})
		if err != nil {
			return err
		}

		if len(apps) == 0 {
			return nil
		}

		sort.Slice(apps, func(i, j int) bool {
			return apps[i].PlayerPeakAllTime > apps[j].PlayerPeakAllTime
		})

		if pack.HasEmptyName() {
			pack.SetName(apps[0].GetName(), false)
		}

		if pack.HasEmptyIcon() {
			pack.Icon = apps[0].GetIcon()
		}

		if pack.ImageLogo == "" {
			pack.ImageLogo = apps[0].GetHeaderImage()
		}

		if pack.ImagePage == "" {
			pack.ImagePage = apps[0].GetHeaderImage()
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
			i64, err = strconv.ParseInt(child.Value, 10, 32)
			pack.BillingType = int32(i64)

		case "licensetype":

			var i64 int64
			i64, err = strconv.ParseInt(child.Value, 10, 32)
			pack.LicenseType = int32(i64)

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
				log.WarnS("More app items", pack.ID)
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
						log.WarnS("More app items2", pack.ID)
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
				log.WarnS("extendedz", pack.ID, child)
			}

		case "extendedasdf": // For package 439981

			log.InfoS(child) // todo

		case "is_available": // For package 439981

			log.InfoS(child) // todo, bool

		case "":

			// Some packages (46028) have blank children

		default:
			log.WarnS(child.Key + " field in package PICS ignored (Package: " + strconv.Itoa(pack.ID) + ")")
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
		steam.WithTimeout(0),
	)

	// ID
	c.OnHTML("h2.pageheader", func(e *colly.HTMLElement) {
		pack.SetName(e.Text, true)
		pack.InStore = true
	})

	//
	c.OnError(func(r *colly.Response, err error) {
		steam.LogSteamError(err)
	})

	err = c.Visit("https://store.steampowered.com/sub/" + strconv.Itoa(pack.ID))
	if err != nil && strings.Contains(err.Error(), "because its not in AllowedDomains") {
		log.InfoS(err)
		return nil
	}

	return err
}

func updatePackageFromStore(pack *mongo.Package) (err error) {

	prices := helpers.ProductPrices{}

	for _, cc := range i18n.GetProdCCs(true) {

		// Get package details
		response, err := steam.GetSteam().GetPackageDetails(uint(pack.ID), cc.ProductCode, steamapi.LanguageEnglish)
		err = steam.AllowSteamCodes(err)
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

				pack.ImageLogo = ""

				_, err := helpers.Head(response.Data.SmallLogo, 0)
				if err != nil {
					return
				}

				pack.ImageLogo = response.Data.SmallLogo
			}()

			wg.Add(1)
			go func() {

				defer wg.Done()

				pack.ImagePage = ""

				code, err := helpers.Head(response.Data.PageImage, time.Second*30)
				if err == helpers.ErrNon200 {
					return
				} else if err != nil {
					log.Err("failed image check", zap.Error(err), zap.String("url", response.Data.PageImage), zap.Int("code", code))
					return
				}

				pack.ImagePage = response.Data.PageImage
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
