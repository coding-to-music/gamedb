package queue

import (
	"regexp"
	"sort"
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
	"github.com/gamedb/gamedb/pkg/helpers/i18n"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	steamHelper "github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/sql/pics"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/gocolly/colly"
	"go.mongodb.org/mongo-driver/bson"
)

type AppMessage struct {
	ID           int                    `json:"id"`
	ChangeNumber int                    `json:"change_number"`
	VDF          map[string]interface{} `json:"vdf"`
}

func appHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := AppMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		var id = payload.ID

		if !helpers.IsValidAppID(id) {
			log.Err(err, payload.ID)
			sendToFailQueue(message)
			continue
		}

		// Load current app
		app, err := mongo.GetApp(id, true)
		if err == mongo.ErrNoDocuments {
			app = mongo.App{}
			app.ID = id
		} else if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			continue
		}

		// Skip if updated in last day, unless its from PICS
		if !config.IsLocal() && !app.ShouldUpdate() && app.ChangeNumber >= payload.ChangeNumber {

			s, err := durationfmt.Format(time.Now().Sub(app.UpdatedAt), "%hh %mm")
			log.Err(err)

			log.Info("Skipping app, updated " + s + " ago")
			message.Ack(false)
			continue
		}

		//
		var appBeforeUpdate = app

		//
		err = updateAppPICS(&app, message, payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToRetryQueue(message)
			continue
		}

		//
		var wg sync.WaitGroup

		// Calls to api.steampowered.com
		var newItems []steamapi.ItemDefArchive
		wg.Add(1)
		go func() {

			defer wg.Done()

			var err error

			newItems, err = updateAppItems(&app)
			if err != nil {
				steamHelper.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		// Calls to store.steampowered.com
		var sales []mongo.Sale
		wg.Add(1)
		go func() {

			defer wg.Done()

			var err error

			err = updateAppDetails(&app)
			if err != nil && err != steamapi.ErrAppNotFound {
				steamHelper.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			sales, err = scrapeApp(&app)
			if err != nil {
				steamHelper.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		// Read from Mongo
		var currentAppItems []int
		wg.Add(1)
		go func() {

			defer wg.Done()

			var err error

			err = updateAppPlaytimeStats(&app)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			currentAppItems, err = getCurrentAppItems(app.ID)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			err = getWishlistCount(&app)
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

		// Save to Mongo
		wg.Add(1)
		go func() {

			defer wg.Done()

			var err error

			err = saveProductPricesToMongo(appBeforeUpdate, app)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			err = saveAppItems(app.ID, newItems, currentAppItems)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			err = saveSales(app, sales)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			err = app.Save()
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

		// Clear caches
		wg.Add(1)
		go func() {

			defer wg.Done()

			var countItem = memcache.FilterToString(bson.D{{"app_id", app.ID}})

			err := memcache.Delete(
				memcache.MemcacheApp(app.ID).Key,
				memcache.MemcacheAppInQueue(app.ID).Key,
				memcache.MemcacheAppTags(app.ID).Key,
				memcache.MemcacheAppCategories(app.ID).Key,
				memcache.MemcacheAppGenres(app.ID).Key,
				memcache.MemcacheAppDemos(app.ID).Key,
				memcache.MemcacheAppRelated(app.ID).Key,
				memcache.MemcacheAppDLC(app.ID).Key,
				memcache.MemcacheAppDevelopers(app.ID).Key,
				memcache.MemcacheAppPublishers(app.ID).Key,
				memcache.MemcacheAppBundles(app.ID).Key,
				memcache.MemcacheAppPackages(app.ID).Key,
				memcache.MemcacheMongoCount(mongo.CollectionAppAchievements.String()+"-"+countItem).Key,
				memcache.MemcacheMongoCount(mongo.CollectionAppItems.String()+"-"+countItem).Key,
				memcache.MemcacheMongoCount(mongo.CollectionAppSales.String()+"-"+countItem).Key,
			)
			if err != nil {
				log.Err(err, id)
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
					log.Err(err, id)
				}
			}
		}()

		// Queue group
		wg.Add(1)
		go func() {

			defer wg.Done()

			var err error

			err = ProduceGroup(GroupMessage{ID: app.GroupID})
			err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
			log.Err(err)
		}()

		wg.Wait()

		if message.ActionTaken {
			continue
		}

		// Produce to sub queues
		var produces = map[rabbit.QueueName]interface{}{
			QueueAppsAchievements: AppAchievementsMessage{ID: app.ID},
			QueueAppsMorelike:     AppMorelikeMessage{ID: app.ID},
			QueueAppsNews:         AppNewsMessage{ID: app.ID},
			QueueAppsSameowners:   AppSameownersMessage{ID: app.ID},
			QueueAppsSteamspy:     AppSteamspyMessage{ID: app.ID},
			QueueAppsTwitch:       AppTwitchMessage{ID: app.ID},
			QueueAppsReviews:      AppReviewsMessage{ID: app.ID},
		}

		for k, v := range produces {
			err = produce(k, v)
			if err != nil {
				log.Err(err)
				sendToRetryQueue(message)
				continue
			}
		}

		//
		message.Ack(false)
	}
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
									log.Info("Missing localization language key", payload.ID, vvvv.Key) // Sometimes the "tokens" map is missing.
								}
							}

						} else {

							for _, vvvv := range vvv.Children {
								localization.RichPresence[vvv.Key].AddToken(vvvv.Key, vvvv.Value)
							}
						}
					}
				} else {
					log.Warning("Missing localization key", payload.ID)
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
				log.Err(err)
			} else {
				app.AlbumMetaData = amd
			}

		default:
			log.Warning(child.Key + " field in app PICS ignored (App: " + strconv.Itoa(app.ID) + ")")
		}
	}

	return nil
}

func updateAppDetails(app *mongo.App) (err error) {

	prices := helpers.ProductPrices{}

	for _, code := range i18n.GetProdCCs(true) {

		var filter []string
		if code.ProductCode != steamapi.ProductCCUS {
			filter = []string{"price_overview"}
		}

		response, b, err := steamHelper.GetSteam().GetAppDetails(uint(app.ID), code.ProductCode, steamapi.LanguageEnglish, filter)
		err = steamHelper.AllowSteamCodes(err, b, nil)
		if err == steamapi.ErrAppNotFound {
			continue
		}
		if err != nil {
			return err
		}

		// Check for missing fields
		go func() {
			err2 := helpers.UnmarshalStrict(b, &map[string]steamapi.AppDetailsBody{})
			if err != nil {
				log.Warning(err2, app.ID)
			}
		}()

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
			app.DLC = response.Data.DLC
			app.DLCCount = len(response.Data.DLC)

			// Packages
			app.Packages = response.Data.Packages

			// Publishers
			gorm, err := sql.GetMySQLClient()
			if err != nil {
				return err
			}

			var publisherIDs []int
			for _, v := range response.Data.Publishers {
				var publisher sql.Publisher
				gorm = gorm.Unscoped().FirstOrCreate(&publisher, sql.Publisher{Name: sql.TrimPublisherName(v)})
				if gorm.Error != nil {
					return gorm.Error
				}
				publisherIDs = append(publisherIDs, publisher.ID)
			}

			app.Publishers = publisherIDs

			// Developers
			gorm, err = sql.GetMySQLClient()
			if err != nil {
				return err
			}

			var developerIDs []int
			for _, v := range response.Data.Developers {
				var developer sql.Developer
				gorm = gorm.Unscoped().FirstOrCreate(&developer, sql.Developer{Name: strings.TrimSpace(v)})
				if gorm.Error != nil {
					return gorm.Error
				}
				developerIDs = append(developerIDs, developer.ID)
			}

			app.Developers = developerIDs

			// Categories
			var categories []int
			for _, v := range response.Data.Categories {
				categories = append(categories, int(v.ID))
			}

			app.Categories = categories

			// Genres
			gorm, err = sql.GetMySQLClient()
			if err != nil {
				return err
			}

			var genreIDs []int
			for _, v := range response.Data.Genres {
				var genre sql.Genre
				gorm = gorm.Unscoped().Assign(sql.Genre{Name: strings.TrimSpace(v.Description)}).FirstOrCreate(&genre, sql.Genre{ID: int(v.ID)})
				if gorm.Error != nil {
					return gorm.Error
				}
				genreIDs = append(genreIDs, genre.ID)
			}

			app.Genres = genreIDs

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

			// Platforms
			app.Platforms = platforms

			// Demos
			var demos []int
			for _, v := range response.Data.Demos {
				demos = append(demos, int(v.AppID))
			}

			app.Demos = demos

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
						log.Err(err, app.ID, assets)
						return
					}

					if _, ok := assetMap["library_hero"]; ok {

						urls := []string{
							"https://steamcdn-a.akamaihd.net/steam/apps/" + strconv.Itoa(app.ID) + "/library_hero.jpg",
							"https://steamcdn-a.akamaihd.net/steam/fpo_apps/" + strconv.Itoa(app.ID) + "/library_hero.jpg",
						}

						for _, u := range urls {
							if helpers.GetResponseCode(u) == 200 {
								app.Background = u
								break
							}
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
			app.ReleaseDate = strings.ToValidUTF8(response.Data.ReleaseDate.Date, "")
			if len(app.ReleaseDate) > 255 {
				app.ReleaseDate = app.ReleaseDate[0:255] // SQL limit
			}
			app.ReleaseDateUnix = helpers.GetReleaseDateUnix(response.Data.ReleaseDate.Date)
			app.ComingSoon = response.Data.ReleaseDate.ComingSoon
		}
	}

	app.Prices = prices

	return nil
}

func updateAppItems(app *mongo.App) (archive []steamapi.ItemDefArchive, err error) {

	meta, _, err := steamHelper.GetSteam().GetItemDefMeta(app.ID)
	if err != nil {
		return archive, err
	}

	if meta.Response.Digest != "" && meta.Response.Digest != app.ItemsDigest {

		archive, _, err = steamHelper.GetSteam().GetItemDefArchive(app.ID, meta.Response.Digest)
		if err != nil {
			return archive, err
		}

		app.Items = len(archive)
		app.ItemsDigest = meta.Response.Digest
	}

	return archive, nil
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
	if helpers.SliceHasString([]string{"media", "movie"}, app.Type) {
		return sales, nil
	}

	var bundleIDs []string
	var tagCounts []mongo.AppTagCount

	// Retry call
	operation := func() (err error) {

		bundleIDs = []string{}

		c := colly.NewCollector(
			colly.URLFilters(appStorePage),
			steamHelper.WithAgeCheckCookie,
			colly.AllowURLRevisit(),
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

						tagCounts = append(tagCounts, mongo.AppTagCount{ID: tagID, Name: match[2], Count: count})
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
					log.Err(app.ID, err)
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
						log.Err(app.ID, err)
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
					log.Err(err, dateString)
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
			steamHelper.LogSteamError(err)
		})

		err = c.Visit("https://store.steampowered.com/app/" + strconv.Itoa(app.ID))
		if err != nil {
			if strings.Contains(err.Error(), "because its not in AllowedDomains") {
				log.Info(err)
				return nil
			}
		}

		return err
	}

	policy := backoff.NewExponentialBackOff()
	policy.InitialInterval = time.Second

	err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Info(err, app.ID) })
	if err != nil {
		return sales, err
	}

	// Save bundle IDs
	var bundleIntIDs = helpers.StringSliceToIntSlice(bundleIDs)

	for _, bundleID := range bundleIntIDs {

		err = ProduceBundle(bundleID)
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		if err != nil {
			log.Err(err)
		}
	}

	// Save tag counts
	sort.Slice(tagCounts, func(i, j int) bool {
		return tagCounts[i].Count > tagCounts[j].Count
	})
	app.TagCounts = tagCounts

	//
	return sales, nil
}

func getCurrentAppItems(appID int) (items []int, err error) {

	resp, err := mongo.GetAppItems(0, 0, bson.D{{"app_id", appID}}, bson.M{"item_def_id": 1})
	for _, v := range resp {
		items = append(items, v.ItemDefID)
	}

	return items, err
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

func getWishlistCount(app *mongo.App) (err error) {

	apps, err := mongo.GetPlayerWishlistAppsByApp(app.ID)
	if err != nil {
		return err
	}

	var count = len(apps)

	app.WishlistCount = count

	var total int
	for _, v := range apps {
		total += v.Order
	}

	if count == 0 {
		app.WishlistAvgPosition = 0
	} else {
		app.WishlistAvgPosition = float64(total) / float64(count)
	}

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

	return mongo.UpdateSales(newSales)
}

func saveAppItems(appID int, newItems []steamapi.ItemDefArchive, currentItemIDs []int) (err error) {

	if len(newItems) == 0 {
		return
	}

	// Make current items map
	var currentItemIDsMap = map[int]bool{}
	for _, v := range currentItemIDs {
		currentItemIDsMap[v] = true
	}

	// Make new items map
	var newItemsMap = map[int]bool{}
	for _, v := range newItems {
		newItemsMap[int(v.ItemDefID)] = true
	}

	// Find new items
	var newDocuments []mongo.AppItem
	for _, v := range newItems {
		_, ok := currentItemIDsMap[int(v.ItemDefID)]
		if !ok {
			appItem := mongo.AppItem{
				AppID:            int(v.AppID),
				Bundle:           v.Bundle,
				Commodity:        bool(v.Commodity),
				DateCreated:      v.DateCreated,
				Description:      v.Description,
				DisplayType:      v.DisplayType,
				DropInterval:     int(v.DropInterval),
				DropMaxPerWindow: int(v.DropMaxPerWindow),
				Hash:             v.Hash,
				IconURL:          v.IconURL,
				IconURLLarge:     v.IconURLLarge,
				ItemDefID:        int(v.ItemDefID),
				ItemQuality:      string(v.ItemQuality),
				Marketable:       bool(v.Marketable),
				Modified:         v.Modified,
				Name:             v.Name,
				Price:            v.Price,
				Promo:            v.Promo,
				Quantity:         int(v.Quantity),
				Timestamp:        v.Timestamp,
				Tradable:         bool(v.Tradable),
				Type:             v.Type,
				WorkshopID:       int64(v.WorkshopID),
				// Exchange:         v.Exchange,
				// Tags:             v.Tags,
			}
			appItem.SetExchange(v.Exchange)
			appItem.SetTags(v.Tags)

			newDocuments = append(newDocuments, appItem)
		}
	}
	err = mongo.UpdateAppItems(newDocuments)
	if err != nil {
		return err
	}

	// Find removed items
	var oldDocumentIDs []int
	for _, v := range currentItemIDs {
		_, ok := newItemsMap[v]
		if !ok {
			oldDocumentIDs = append(oldDocumentIDs, v)
		}
	}

	return mongo.DeleteAppItems(appID, oldDocumentIDs)
}
