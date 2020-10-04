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
	"github.com/cenkalti/backoff/v4"
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

type AppMessage struct {
	ID           int                    `json:"id"`
	ChangeNumber int                    `json:"change_number"`
	VDF          map[string]interface{} `json:"vdf"`
}

func appHandler(message *rabbit.Message) {

	payload := AppMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToFailQueue(message)
		return
	}

	var id = payload.ID

	if !helpers.IsValidAppID(id) {
		log.ErrS(err, payload.ID)
		sendToFailQueue(message)
		return
	}

	// Load current app
	app, err := mongo.GetApp(id, true)
	if err == mongo.ErrNoDocuments {
		app = mongo.App{}
		app.ID = id
	} else if err != nil {
		log.ErrS(err, payload.ID)
		sendToRetryQueue(message)
		return
	}

	// Skip if updated in last day, unless its from PICS
	if !config.IsLocal() && !app.ShouldUpdate() && app.ChangeNumber >= payload.ChangeNumber {

		s, err := durationfmt.Format(time.Since(app.UpdatedAt), "%hh %mm")
		if err != nil {
			log.ErrS(err)
		}

		log.InfoS("Skipping app, updated " + s + " ago")
		message.Ack()
		return
	}

	//
	var appBeforeUpdate = app

	//
	err = updateAppPICS(&app, message, payload)
	if err != nil {
		log.Err(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	//
	var wg sync.WaitGroup

	// Calls to store.steampowered.com
	var sales []mongo.Sale
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		err = updateAppDetails(&app)
		if err != nil && err != steamapi.ErrAppNotFound {
			steam.LogSteamError(err, payload.ID)
			sendToRetryQueue(message)
			return
		}

		sales, err = scrapeApp(&app)
		if err != nil {
			steam.LogSteamError(err, payload.ID)
			sendToRetryQueue(message)
			return
		}
	}()

	// Read from Mongo
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		err = updateAppPlaytimeStats(&app)
		if err != nil {
			log.ErrS(err, payload.ID)
			sendToRetryQueue(message)
			return
		}

		err = updateAppOwners(&app)
		if err != nil {
			log.ErrS(err, payload.ID)
			sendToRetryQueue(message)
			return
		}

		err = updateAppBadgeOwners(&app)
		if err != nil {
			log.ErrS(err, payload.ID)
			sendToRetryQueue(message)
			return
		}

		err = updateAppCountries(&app)
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

	// Save to Mongo
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		err = saveProductPricesToMongo(appBeforeUpdate, app)
		if err != nil {
			log.ErrS(err, payload.ID)
			sendToRetryQueue(message)
			return
		}

		err = saveSales(app, sales)
		if err != nil {
			log.ErrS(err, payload.ID)
			sendToRetryQueue(message)
			return
		}

		err = replaceAppRow(app)
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

	// Clear caches
	wg.Add(1)
	go func() {

		defer wg.Done()

		var items = []string{
			// memcache.MemcacheApp(app.ID).Key, // Done in replaceAppRow
			memcache.MemcacheAppInQueue(app.ID).Key,
			memcache.MemcacheAppStats(mongo.StatsTypeTags.String(), app.ID).Key,
			memcache.MemcacheAppStats(mongo.StatsTypeCategories.String(), app.ID).Key,
			memcache.MemcacheAppStats(mongo.StatsTypeGenres.String(), app.ID).Key,
			memcache.MemcacheAppStats(mongo.StatsTypeDevelopers.String(), app.ID).Key,
			memcache.MemcacheAppStats(mongo.StatsTypePublishers.String(), app.ID).Key,
			memcache.MemcacheAppDemos(app.ID).Key,
			memcache.MemcacheAppRelated(app.ID).Key,
			memcache.MemcacheAppBundles(app.ID).Key,
			memcache.MemcacheAppPackages(app.ID).Key,
			memcache.MemcacheAppNoAchievements(app.ID).Key,
			memcache.MemcacheMongoCount(mongo.CollectionAppSales.String(), bson.D{{"app_id", app.ID}}).Key,
		}

		err := memcache.Delete(items...)
		if err != nil {
			log.ErrS(err, payload.ID)
			sendToRetryQueue(message)
			return
		}
	}()

	// Send websocket
	wg.Add(1)
	go func() {

		defer wg.Done()

		if payload.ChangeNumber > 0 {

			var err error

			wsPayload := IntPayload{ID: id}
			err = ProduceWebsocket(wsPayload, websockets.PageApp)
			if err != nil {
				log.ErrS(err, id)
			}
		}
	}()

	wg.Wait()

	if message.ActionTaken {
		return
	}

	// Produce to sub queues
	var produces = []QueueMessageInterface{
		// AppsSearchMessage{App: app}, // Done in sub queues
		AppAchievementsMessage{AppID: app.ID, AppName: app.Name, AppOwners: app.Owners},
		AppMorelikeMessage{AppID: app.ID},
		AppNewsMessage{AppID: app.ID},
		AppSameownersMessage{AppID: app.ID},
		AppTwitchMessage{AppID: app.ID},
		AppReviewsMessage{AppID: app.ID},
		AppItemsMessage{AppID: app.ID, OldDigect: app.ItemsDigest},
	}

	if app.GroupID == "" {
		produces = append(produces, FindGroupMessage{AppID: app.ID})
	} else {
		produces = append(produces, GroupMessage{ID: app.GroupID})
	}

	for _, v := range produces {
		err = produce(v.Queue(), v)
		if err != nil {
			log.ErrS(err)
			sendToRetryQueue(message)
			break
		}
	}

	if message.ActionTaken {
		return
	}

	//
	message.Ack()
}

func updateAppPICS(app *mongo.App, message *rabbit.Message, payload AppMessage) (err error) {

	if !config.IsLocal() {
		if payload.ChangeNumber == 0 || app.ChangeNumber >= payload.ChangeNumber {
			return nil
		}
	}

	var kv = steamvdf.FromMap(payload.VDF)

	app.ID = payload.ID

	if payload.ChangeNumber > 0 {
		app.ChangeNumber = payload.ChangeNumber
		app.ChangeNumberDate = message.FirstSeen()
	}

	// Reset values that might be removed
	app.Common = map[string]string{}
	app.Tags = []int{}
	app.Extended = map[string]string{}
	app.Config = map[string]string{}
	app.Launch = []pics.PICSAppConfigLaunchItem{}
	app.Depots = pics.Depots{}
	app.PublicOnly = false
	app.UFS = map[string]string{}
	app.Install = map[string]interface{}{}
	app.Localization = pics.Localisation{}
	app.SystemRequirements = map[string]interface{}{}

	if len(kv.Children) == 1 && kv.Children[0].Key == "appinfo" {
		kv = kv.Children[0]
	}

	if len(kv.Children) == 0 {
		return nil
	}

	for _, child := range kv.Children {

		switch child.Key {
		case "appid":
			//
		case "common":

			var common = pics.PICSKeyValues{}
			var tags []int

			for _, vv := range child.Children {

				if vv.Key == "store_tags" {
					stringTags := vv.GetChildrenAsSlice()
					tags = helpers.StringSliceToIntSlice(stringTags)
				}

				if vv.Key == "name" {
					name := strings.TrimSpace(vv.Value)
					if name != "" {
						app.Name = name
					}
				}

				if vv.Key == "type" {
					name := strings.TrimSpace(vv.Value)
					if name != "" {
						app.Type = strings.ToLower(name) // Has priority over store API
					}
				}

				if vv.Key == "icon" {
					icon := strings.TrimSpace(vv.Value)
					if icon != "" {
						app.Icon = icon
					}
				}

				if vv.Key == "steam_release_date" {
					i, err := strconv.ParseInt(vv.Value, 10, 64)
					if err == nil && i > 0 {
						app.ReleaseDateUnix = i
					}
				}

				if vv.Key == "original_release_date" {
					i, err := strconv.ParseInt(vv.Value, 10, 64)
					if err == nil && i > 0 {
						app.ReleaseDateOriginal = i
					}
				}

				common[vv.Key] = vv.String()
			}

			app.Common = common
			app.Tags = tags

		case "extended":

			app.Extended = child.GetChildrenAsMap()

		case "config":

			c, launch := getAppConfig(child)

			app.Config = c
			app.Launch = launch

		case "depots":

			app.Depots = getAppDepots(child)

		case "public_only":

			if child.Value == "1" {
				app.PublicOnly = true
			}

		case "ufs":

			var ufs = pics.PICSKeyValues{}
			for _, vv := range child.Children {
				ufs[vv.Key] = vv.String()
			}

			app.UFS = ufs

		case "install":

			app.Install = child.ToMapInner()

		case "localization":

			localization := pics.Localisation{}
			for _, vv := range child.Children {
				if vv.Key == "richpresence" {
					for _, vvv := range vv.Children {

						localization.AddLanguage(vvv.Key, &pics.LocalisationLanguage{})

						// Some apps (657730, 1025600) skip the `tokens` level
						if vvv.HasChild("tokens") {

							for _, vvvv := range vvv.Children {
								if vvvv.Key == "tokens" {
									for _, vvvvv := range vvvv.Children {
										localization.RichPresence[vvv.Key].AddToken(vvvvv.Key, vvvvv.Value)
									}
								} else {
									log.InfoS("Missing localization language key", payload.ID, vvvv.Key) // Sometimes the "tokens" map is missing.
								}
							}

						} else {

							for _, vvvv := range vvv.Children {
								localization.RichPresence[vvv.Key].AddToken(vvvv.Key, vvvv.Value)
							}
						}
					}
				} else {
					log.WarnS("Missing localization key", payload.ID)
				}
			}

			app.Localization = localization
			app.LocalizationCount = len(localization.RichPresence)

		case "sysreqs":

			app.SystemRequirements = child.ToMapInner()

		case "albummetadata":

			amd := pics.AlbumMetaData{}
			err = helpers.MarshalUnmarshal(child.ToMapInner(), &amd)
			if err != nil {
				log.ErrS(err)
			} else {
				app.AlbumMetaData = amd
			}

		default:
			log.Warn(child.Key + " field in app PICS ignored (App: " + strconv.Itoa(app.ID) + ")")
		}
	}

	return nil
}

func updateAppDetails(app *mongo.App) (err error) {

	prices := helpers.ProductPrices{}

	for _, code := range i18n.GetProdCCs(true) {

		// No price_overview filter so we can get `is_free`
		response, err := steam.GetSteam().GetAppDetails(uint(app.ID), code.ProductCode, steamapi.LanguageEnglish, nil)
		err = steam.AllowSteamCodes(err)

		// Not available in language
		if err == steamapi.ErrAppNotFound || response.Data == nil {
			continue
		}

		// Retry app
		if err != nil {
			return err
		}

		//
		prices.AddPriceFromApp(code.ProductCode, response)

		if code.ProductCode == steamapi.ProductCCUS {

			// Screenshots
			var images []helpers.AppImage
			for _, v := range response.Data.Screenshots {
				images = append(images, helpers.AppImage{
					PathFull:      v.PathFull,
					PathThumbnail: v.PathThumbnail,
				})
			}

			app.Screenshots = images

			// Movies
			var videos []helpers.AppVideo
			for _, v := range response.Data.Movies {
				videos = append(videos, helpers.AppVideo{
					PathFull:      v.Webm.Max,
					PathThumbnail: v.Thumbnail,
					Title:         v.Name,
				})
			}

			app.Movies = videos

			// DLC
			app.DLCCount = len(response.Data.DLC)

			err = ProduceDLC(app.ID, response.Data.DLC)
			if err != nil {
				log.ErrS(err)
			}

			// Packages
			app.Packages = response.Data.Packages

			// Publishers
			app.Publishers, err = mongo.FindOrCreateStatsByName(mongo.StatsTypePublishers, response.Data.Publishers)
			if err != nil {
				return err
			}

			// Developers
			app.Developers, err = mongo.FindOrCreateStatsByName(mongo.StatsTypeDevelopers, response.Data.Developers)
			if err != nil {
				return err
			}

			// Genres
			app.Genres = response.Data.Genres.IDs()

			err = mongo.EnsureStat(mongo.StatsTypeGenres, response.Data.Genres.IDs(), response.Data.Genres.Names())
			if err != nil {
				return err
			}

			// Categories
			app.Categories = response.Data.Categories.IDs()

			err = mongo.EnsureStat(mongo.StatsTypeCategories, response.Data.Categories.IDs(), response.Data.Categories.Names())
			if err != nil {
				return err
			}

			// Demos
			for _, v := range response.Data.Demos {
				app.Demos = append(app.Demos, int(v.AppID))
			}

			// Platforms
			var platforms []string
			if response.Data.Platforms.Linux {
				platforms = append(platforms, mongo.PlatformLinux)
			}
			if response.Data.Platforms.Windows {
				platforms = append(platforms, mongo.PlatformWindows)
			}
			if response.Data.Platforms.Mac {
				platforms = append(platforms, mongo.PlatformMac)
			}

			app.Platforms = platforms

			// Images
			var wg sync.WaitGroup

			// Save background image
			wg.Add(1)
			go func() {

				defer wg.Done()

				app.Background = ""

				if assets, ok := app.Common["library_assets"]; ok {

					assetMap := map[string]interface{}{}
					err := helpers.Unmarshal([]byte(assets), &assetMap)
					if err != nil {
						log.ErrS(err, app.ID, assets)
						return
					}

					if _, ok := assetMap["library_hero"]; ok {

						urls := []string{
							"https://steamcdn-a.akamaihd.net/steam/apps/" + strconv.Itoa(app.ID) + "/library_hero.jpg",
							"https://steamcdn-a.akamaihd.net/steam/fpo_apps/" + strconv.Itoa(app.ID) + "/library_hero.jpg",
						}

						for _, u := range urls {
							_, err := helpers.Head(u, 0)
							if err != nil {
								continue
							}
							app.Background = u
							break
						}
					}
				}
			}()

			wg.Wait()

			// Other
			if app.Name == "" && response.Data.Name != "" {
				app.Name = strings.TrimSpace(response.Data.Name)
			}

			// Getting type from PICS has priority
			if app.Type == "" {
				app.Type = strings.ToLower(response.Data.Type)
			}

			app.IsFree = response.Data.IsFree
			app.ShortDescription = response.Data.ShortDescription
			app.MetacriticScore = response.Data.Metacritic.Score
			app.MetacriticURL = response.Data.Metacritic.URL
			app.GameID = int(response.Data.Fullgame.AppID)
			app.GameName = response.Data.Fullgame.Name
			app.ComingSoon = response.Data.ReleaseDate.ComingSoon

			// Release dates
			releaseDate := strings.ToValidUTF8(response.Data.ReleaseDate.Date, "")

			if len(releaseDate) > 255 {
				releaseDate = releaseDate[0:255] // Column limit
			}

			if app.ReleaseDate == "" {
				app.ReleaseDate = releaseDate
			}

			if app.ReleaseDateUnix == 0 {
				app.ReleaseDateUnix = helpers.GetReleaseDateUnix(releaseDate)
			}
		}
	}

	app.Prices = prices

	return nil
}

//noinspection RegExpRedundantEscape
var (
	appStorePage     = regexp.MustCompile(`store\.steampowered\.com/app/[0-9]+$`)
	appStorePageTags = regexp.MustCompile(`\{"tagid":([0-9]+),"name":"([a-zA-Z0-9-&'. ]+)","count":([0-9]+),"browseable":[a-z]{4,5}}`)
)

func scrapeApp(app *mongo.App) (sales []mongo.Sale, err error) {

	// This app causes infinite redirects..
	if app.ID == 12820 {
		return sales, nil
	}

	// Skip these app types
	if helpers.SliceHasString(app.Type, []string{"media", "movie"}) {
		return sales, nil
	}

	var bundleIDs []string

	// Retry call
	operation := func() (err error) {

		bundleIDs = []string{}

		c := colly.NewCollector(
			colly.URLFilters(appStorePage),
			colly.AllowURLRevisit(),
			steam.WithAgeCheckCookie,
			steam.WithTimeout(0),
		)

		// Tags
		c.OnHTML("script", func(e *colly.HTMLElement) {
			matches := appStorePageTags.FindAllStringSubmatch(e.Text, -1)
			if len(matches) > 0 {
				for _, match := range matches {
					if len(match) == 4 {

						tagID, err := strconv.Atoi(match[1])
						if err != nil {
							continue
						}

						count, err := strconv.Atoi(match[3])
						if err != nil {
							continue
						}

						var tagFound bool
						for k, tag := range app.TagCounts {
							if tag.ID == tagID {
								tagFound = true
								app.TagCounts[k].Name = match[2]
								app.TagCounts[k].Count = count
							}
						}

						if !tagFound {
							app.TagCounts = append(app.TagCounts, mongo.AppTagCount{
								ID:    tagID,
								Name:  match[2],
								Count: count,
							})
						}
					}
				}
			}
		})

		// Bundles
		c.OnHTML("div.game_area_purchase_game_wrapper input[name=bundleid]", func(e *colly.HTMLElement) {
			bundleIDs = append(bundleIDs, e.Attr("value"))
		})

		// Sales
		var i = 0
		c.OnHTML("div.game_area_purchase_game p", func(e *colly.HTMLElement) {

			h1 := strings.TrimSuffix(e.DOM.Parent().Find("h1").Text(), "Buy ")

			var sale = mongo.Sale{
				AppID:    app.ID,
				SubOrder: i,
				SaleName: h1,
			}

			// Set discount percent
			discountText := e.DOM.Parent().Find("div.discount_pct").Text()
			if discountText != "" {
				sale.SalePercent, err = strconv.Atoi(helpers.RegexNonNumbers.ReplaceAllString(discountText, ""))
				if err != nil {
					log.ErrS(app.ID, err)
				}
			}

			// Get sub ID
			subIDString, exists := e.DOM.Parent().Find("input[name=subid]").Attr("value")
			if exists {
				subID, err := strconv.Atoi(subIDString)
				if err == nil {
					sale.SubID = subID
				}
			}

			if strings.Contains(e.Text, "Offer ends in") {

				// DAILY DEAL! Offer ends in <span id=348647_countdown_0></span>

				// Get type
				index := strings.Index(e.Text, "Offer ends in")
				sale.SaleType = strings.ToLower(strings.Trim(e.Text[:index], " !"))

				// Get end time
				ts := helpers.RegexTimestamps.FindString(e.DOM.Parent().Text())
				if ts != "" {
					t, err := strconv.ParseInt(ts, 10, 64)
					if err != nil {
						log.ErrS(app.ID, err)
					} else {
						sale.SaleEnd = time.Unix(t, 0)
					}
				}

				sales = append(sales, sale)
				i++

			} else if strings.Contains(e.Text, "Offer ends") {

				// SPECIAL PROMOTION! Offer ends 29 August

				sale.SaleEndEstimate = true

				// Get type
				index := strings.Index(e.Text, "Offer ends")
				sale.SaleType = strings.ToLower(strings.Trim(e.Text[:index], " !"))

				// Get end time
				dateString := strings.TrimSpace(e.Text[index+len("Offer ends"):])

				var t time.Time
				var timeSet bool

				if !timeSet {
					t, err = time.Parse("2 January", dateString)
					if err == nil {
						timeSet = true
					}
				}

				if !timeSet {
					t, err = time.Parse("January 2", dateString)
					if err == nil {
						timeSet = true
					}
				}

				if timeSet {

					now := time.Now()

					t = t.AddDate(now.Year(), 0, 0)
					if t.Before(now) {
						t = t.AddDate(1, 0, 0)
					}

					sale.SaleEnd = t

				} else {
					log.ErrS(err, dateString)
					return
				}

				sales = append(sales, sale)
				i++

			} else if strings.Contains(e.Text, "Ends in") {

				// Play for free! Ends in 2 days

				sale.SaleEndEstimate = true
				sale.SaleEnd = time.Now()

				// Get type
				index := strings.Index(e.Text, "Ends in")
				sale.SaleType = strings.ToLower(strings.Trim(e.Text[:index], " !"))

				// Get end time
				dateString := strings.TrimSpace(e.Text[index+len("Ends in"):])

				daysRegex := regexp.MustCompile("([0-9]{1,2}) day")
				daysMatches := daysRegex.FindStringSubmatch(dateString)
				if len(daysMatches) == 2 {
					i, err := strconv.Atoi(daysMatches[1])
					if err == nil {
						sale.SaleEnd = sale.SaleEnd.AddDate(0, 0, i)
					}
				}

				hoursRegex := regexp.MustCompile("([0-9]{1,2}) day")
				hoursMatches := hoursRegex.FindStringSubmatch(dateString)
				if len(hoursMatches) == 2 {
					i, err := strconv.Atoi(hoursMatches[1])
					if err == nil {
						sale.SaleEnd = sale.SaleEnd.Add(time.Hour * time.Duration(i))
					}
				}

				sales = append(sales, sale)
				i++
			}
		})

		//
		c.OnError(func(r *colly.Response, err error) {
			steam.LogSteamError(err)
		})

		err = c.Visit("https://store.steampowered.com/app/" + strconv.Itoa(app.ID))
		if err != nil {
			if strings.Contains(err.Error(), "because its not in AllowedDomains") {
				log.InfoS(err)
				return nil
			}
		}

		return err
	}

	policy := backoff.NewExponentialBackOff()
	policy.InitialInterval = time.Second

	err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.InfoS(err, app.ID) })
	if err != nil {
		return sales, err
	}

	// Save bundle IDs
	var bundleIntIDs = helpers.StringSliceToIntSlice(bundleIDs)

	for _, bundleID := range bundleIntIDs {

		err = ProduceBundle(bundleID)
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		if err != nil {
			log.ErrS(err)
		}
	}

	//
	return sales, nil
}

func updateAppPlaytimeStats(app *mongo.App) (err error) {

	// Playtime
	players, err := mongo.GetAppPlayTimes(app.ID)
	if err != nil {
		return err
	}

	if len(players) == 0 {

		app.PlaytimeTotal = 0
		app.PlaytimeAverage = 0

	} else {

		var minutes int64
		for _, v := range players {
			minutes += int64(v.AppTime)
		}

		app.PlaytimeTotal = minutes
		app.PlaytimeAverage = float64(minutes) / float64(len(players))
	}

	return nil
}

func updateAppOwners(app *mongo.App) (err error) {

	app.Owners, err = mongo.CountDocuments(mongo.CollectionPlayerApps, bson.D{{"app_id", app.ID}}, 0)
	return err
}

func updateAppBadgeOwners(app *mongo.App) (err error) {

	filter := bson.D{
		{"app_id", app.ID},
		{"badge_id", 0},
		{"badge_foil", false},
	}

	app.BadgeOwners, err = mongo.CountDocuments(mongo.CollectionPlayerBadges, filter, 0)
	return err
}

func updateAppCountries(app *mongo.App) (err error) {

	countries, err := mongo.GetAppPlayersByCountry(app.ID)
	if err != nil {
		return err
	}

	m := map[string]int{}
	for _, v := range countries {
		m[v.Country] = v.Count
	}

	app.Countries = m

	return nil
}

func saveSales(app mongo.App, newSales []mongo.Sale) (err error) {

	// Get current app sales
	oldSales, err := mongo.GetAppSales(app.ID)
	if err != nil {
		return err
	}

	var oldSalesMap = map[string]mongo.Sale{}
	for _, v := range oldSales {
		oldSalesMap[v.GetKey()] = v
	}

	for k, v := range newSales {

		newSales[k].AppName = app.GetName()
		newSales[k].AppIcon = app.GetIcon()
		newSales[k].AppLowestPrice = map[steamapi.ProductCC]int{} // todo
		newSales[k].AppRating = app.ReviewsScore
		newSales[k].AppReleaseDate = time.Unix(app.ReleaseDateUnix, 0)
		newSales[k].AppReleaseDateString = app.ReleaseDate
		newSales[k].AppPlayersWeek = app.PlayerPeakWeek
		newSales[k].AppTags = app.Tags
		newSales[k].AppPlatforms = app.Platforms
		newSales[k].AppCategories = app.Categories
		newSales[k].AppType = app.Type

		if val, ok := oldSalesMap[v.GetKey()]; ok {
			newSales[k].SaleStart = val.SaleStart
		} else {
			newSales[k].SaleStart = time.Now()
		}

		newSales[k].AppPrices = app.GetPrices().Map()
	}

	return mongo.ReplaceSales(newSales)
}

func replaceAppRow(app mongo.App) (err error) {

	_, err = mongo.ReplaceOne(mongo.CollectionApps, bson.D{{"_id", app.ID}}, app)
	if err != nil {
		return err
	}

	// Cache cleared here to stop any race conditions with other queues
	return memcache.Delete(memcache.MemcacheApp(app.ID).Key)
}
