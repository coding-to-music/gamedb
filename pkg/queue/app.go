package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/cenkalti/backoff"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/sql/pics"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/gocolly/colly"
	influx "github.com/influxdata/influxdb1-client"
	"github.com/mitchellh/mapstructure"
	"github.com/nicklaw5/helix"
	"github.com/streadway/amqp"
)

type appMessage struct {
	ID          int                  `json:"id"`
	PICSAppInfo rabbitMessageProduct `json:"PICSAppInfo"`
}

type appQueue struct {
	baseQueue
}

func (q appQueue) processMessages(msgs []amqp.Delivery) {

	msg := msgs[0]

	var err error
	var payload = baseMessage{
		Message:       appMessage{},
		OriginalQueue: queueGoApps,
	}

	err = helpers.Unmarshal(msg.Body, &payload)
	if err != nil {
		logError(err)
		payload.ack(msg)
		return
	}

	var message appMessage
	err = mapstructure.Decode(payload.Message, &message)
	if err != nil {
		logError(err)
		payload.ack(msg)
		return
	}

	if payload.Attempt > 1 {
		logInfo("Consuming app " + strconv.Itoa(message.ID) + ", attempt " + strconv.Itoa(payload.Attempt))
	}

	if !helpers.IsValidAppID(message.ID) {
		logError(errors.New("invalid app ID: " + strconv.Itoa(message.ID)))
		payload.ack(msg)
		return
	}

	// Load current app
	gorm, err := sql.GetMySQLClient()
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	app := sql.App{}
	gorm = gorm.FirstOrInit(&app, sql.App{ID: message.ID})
	if gorm.Error != nil {
		logError(gorm.Error, message.ID)
		payload.ackRetry(msg)
		return
	}

	var newApp bool
	if app.CreatedAt.IsZero() {
		newApp = true
	}

	// Skip if updated in last day, unless its from PICS
	if !config.IsLocal() {
		if app.UpdatedAt.Unix() > time.Now().Add(time.Hour * 24 * -1).Unix() {
			if app.ChangeNumber >= message.PICSAppInfo.ChangeNumber {
				logInfo("Skipping app, updated in last day")
				payload.ack(msg)
				return
			}
		}
	}

	//
	var appBeforeUpdate = app

	//
	err = updateAppPICS(&app, payload, message)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	//
	var wg sync.WaitGroup

	// Calls to api.steampowered.com
	var newItems []steam.ItemDefArchive
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		schema, err := updateAppSchema(&app)
		if err != nil {
			helpers.LogSteamError(err, message.ID)
			payload.ackRetry(msg)
			return
		}

		err = updateAppAchievements(&app, schema)
		if err != nil {
			helpers.LogSteamError(err, message.ID)
			payload.ackRetry(msg)
			return
		}

		err = updateAppNews(&app)
		if err != nil {
			helpers.LogSteamError(err, message.ID)
			payload.ackRetry(msg)
			return
		}

		newItems, err = updateAppItems(&app)
		if err != nil {
			helpers.LogSteamError(err, message.ID)
			payload.ackRetry(msg)
			return
		}
	}()

	// Calls to store.steampowered.com
	var offers []mongo.Offer
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		offers, err = updateAppDetails(&app)
		if err != nil && err != steam.ErrAppNotFound {
			helpers.LogSteamError(err, message.ID)
			payload.ackRetry(msg)
			return
		}

		err = updateAppReviews(&app)
		if err != nil {
			helpers.LogSteamError(err, message.ID)
			payload.ackRetry(msg)
			return
		}

		err = updateBundlesAndOffers(&app)
		if err != nil {
			helpers.LogSteamError(err, message.ID)
			payload.ackRetry(msg)
			return
		}
	}()

	// Calls to steamspy.com
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		err = updateAppSteamSpy(&app)
		if err != nil {
			logInfo(err, message.ID)
		}
	}()

	// Calls to Twitch
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		err = updateAppTwitch(&app)
		if err != nil {
			logError(err, message.ID)
			payload.ackRetry(msg)
			return
		}
	}()

	// Calls to Mongo
	var currentAppItems []int
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		err = updateAppPlaytimeStats(&app)
		if err != nil {
			logError(err, message.ID)
			payload.ackRetry(msg)
			return
		}

		currentAppItems, err = getCurrentAppItems(app.ID)
		if err != nil {
			logError(err, message.ID)
			payload.ackRetry(msg)
			return
		}
	}()

	wg.Wait()

	if payload.actionTaken {
		return
	}

	// Save prices to Mongo
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		err = savePriceChanges(appBeforeUpdate, app)
		if err != nil {
			logError(err, message.ID)
			payload.ackRetry(msg)
			return
		}

		err = saveAppItems(app.ID, newItems, currentAppItems)
		if err != nil {
			logError(err, message.ID)
			payload.ackRetry(msg)
			return
		}

		err = saveOffers(appBeforeUpdate, offers)
		if err != nil {
			logError(err, message.ID)
			payload.ackRetry(msg)
			return
		}
	}()

	// Save app to MySQL
	wg.Add(1)
	go func() {

		defer wg.Done()

		app.Type = strings.ToLower(app.Type)
		app.ReleaseState = strings.ToLower(app.ReleaseState)

		gorm = gorm.Save(&app)
		if gorm.Error != nil {
			logError(gorm.Error, message.ID)
			payload.ackRetry(msg)
			return
		}
	}()

	// Save app score etc to Influx
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		err = saveAppToInflux(app)
		if err != nil {
			logError(err, message.ID)
			payload.ackRetry(msg)
			return
		}
	}()

	wg.Wait()

	if payload.actionTaken {
		return
	}

	// Clear caches
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		if app.ReleaseDateUnix > time.Now().Unix() && newApp {

			err = helpers.RemoveKeyFromMemCacheViaPubSub(helpers.MemcacheUpcomingAppsCount.Key)
			logError(err, message.ID)
		}
	}()

	// Send websocket
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		wsPayload := websockets.PubSubIDPayload{}
		wsPayload.ID = message.ID
		wsPayload.Pages = []websockets.WebsocketPage{websockets.PageApp}

		_, err = helpers.Publish(helpers.PubSubTopicWebsockets, wsPayload)
		if err != nil {
			logError(err, message.ID)
		}
	}()

	wg.Wait()

	if payload.actionTaken {
		return
	}

	//
	payload.ack(msg)
}

func updateAppPICS(app *sql.App, payload baseMessage, message appMessage) (err error) {

	if message.PICSAppInfo.ID == 0 {
		return nil
	}

	if app.ChangeNumber > message.PICSAppInfo.ChangeNumber {
		return nil
	}

	app.ID = message.ID
	app.ChangeNumber = message.PICSAppInfo.ChangeNumber
	app.ChangeNumberDate = payload.FirstSeen

	// Reset values that might be removed
	app.Common = ""
	app.Tags = ""
	app.Extended = ""
	app.Config = ""
	app.Launch = ""
	app.Depots = ""
	app.PublicOnly = false
	app.UFS = ""
	app.Install = ""
	app.Localization = ""
	app.SystemRequirements = ""

	for _, v := range message.PICSAppInfo.KeyValues.Children {

		switch v.Name {
		case "appid":

			// No need for this
			// var i64 int64
			// i64, err = strconv.ParseInt(v.Value.(string), 10, 32)
			// if err != nil {
			// 	return err
			// }
			// app.ID = int(i64)

		case "common":

			var common = pics.PICSKeyValues{}
			var tags []int

			for _, vv := range v.Children {

				if vv.Name == "store_tags" {
					stringTags := vv.GetChildrenAsSlice()
					tags = helpers.StringSliceToIntSlice(stringTags)
				}
				if vv.Name == "name" {
					name := strings.TrimSpace(vv.Value.(string))
					if name != "" {
						app.Name = name
					}
				}

				common[vv.Name] = vv.String()
			}

			b, err := json.Marshal(common)
			if err != nil {
				return err
			}
			app.Common = string(b)

			b, err = json.Marshal(tags)
			if err != nil {
				return err
			}
			app.Tags = string(b)

		case "extended":

			b, err := json.Marshal(v.GetExtended())
			if err != nil {
				return err
			}

			app.Extended = string(b)

		case "config":

			c, launch := v.GetAppConfig()

			b, err := json.Marshal(c)
			if err != nil {
				return err
			}
			app.Config = string(b)

			b, err = json.Marshal(launch)
			if err != nil {
				return err
			}
			app.Launch = string(b)

		case "depots":

			b, err := json.Marshal(v.GetAppDepots())
			if err != nil {
				return err
			}

			app.Depots = string(b)

		case "public_only":

			if v.Value.(string) == "1" {
				app.PublicOnly = true
			}

		case "ufs":

			var ufs = pics.PICSKeyValues{}
			for _, vv := range v.Children {
				ufs[vv.Name] = vv.String()
			}

			b, err := json.Marshal(ufs)
			if err != nil {
				return err
			}

			app.UFS = string(b)

		case "install":

			b, err := json.Marshal(v.ToNestedMaps())
			if err != nil {
				return err
			}

			app.Install = string(b)

		case "localization":

			localization := pics.Localisation{}
			for _, vv := range v.Children {
				if vv.Name == "richpresence" {
					for _, vvv := range vv.Children {
						localization.AddLanguage(vvv.Name, &pics.LocalisationLanguage{})
						for _, vvvv := range vvv.Children {
							if vvvv.Name == "tokens" {
								for _, vvvvv := range vvvv.Children {
									localization.RichPresence[vvv.Name].AddToken(vvvvv.Name, vvvvv.Value.(string))
								}
							} else {
								log.Warning("Missing localization language key", message.ID, vvvv.Name)
							}
						}
					}
				} else {
					log.Warning("Missing localization key", message.ID)
				}
			}

			b, err := json.Marshal(localization)
			if err != nil {
				return err
			}

			app.Localization = string(b)

		case "sysreqs":

			b, err := json.MarshalIndent(v.ToNestedMaps(), "", "  ")
			if err != nil {
				return err
			}

			app.SystemRequirements = string(b)

		default:
			logWarning(v.Name + " field in app PICS ignored (Change " + strconv.Itoa(app.ChangeNumber) + ")")
		}
	}

	return nil
}

func updateAppDetails(app *sql.App) (offers []mongo.Offer, err error) {

	prices := sql.ProductPrices{}

	for _, code := range helpers.GetProdCCs(true) {

		var filter []string
		if code.ProductCode != steam.ProductCCUS {
			filter = []string{"price_overview"}
		}

		response, b, err := helpers.GetSteam().GetAppDetails(app.ID, code.ProductCode, steam.LanguageEnglish, filter)
		err = helpers.AllowSteamCodes(err, b, nil)
		if err == steam.ErrAppNotFound {
			continue
		}
		if err != nil {
			return offers, err
		}

		// Check for missing fields
		go func() {
			err2 := helpers.UnmarshalStrict(b, &map[string]steam.AppDetailsBody{})
			if err != nil {
				log.Warning(err2, app.ID)
			}
		}()

		//
		prices.AddPriceFromApp(code.ProductCode, response)

		if code.ProductCode == steam.ProductCCUS {

			// Screenshots
			var images []sql.AppImage
			for _, v := range response.Data.Screenshots {
				images = append(images, sql.AppImage{
					PathFull:      v.PathFull,
					PathThumbnail: v.PathThumbnail,
				})
			}

			b, err := json.Marshal(images)
			if err != nil {
				return offers, err
			}

			app.Screenshots = string(b)

			// Movies
			var videos []sql.AppVideo
			for _, v := range response.Data.Movies {
				videos = append(videos, sql.AppVideo{
					PathFull:      v.Webm.Max,
					PathThumbnail: v.Thumbnail,
					Title:         v.Name,
				})
			}

			b, err = json.Marshal(videos)
			if err != nil {
				return offers, err
			}

			app.Movies = string(b)

			// DLC
			b, err = json.Marshal(response.Data.DLC)
			if err != nil {
				return offers, err
			}

			app.DLC = string(b)
			app.DLCCount = len(response.Data.DLC)

			// Packages
			b, err = json.Marshal(response.Data.Packages)
			if err != nil {
				return offers, err
			}

			app.Packages = string(b)

			// Publishers
			gorm, err := sql.GetMySQLClient()
			if err != nil {
				return offers, err
			}

			var publisherIDs []int
			for _, v := range response.Data.Publishers {
				var publisher sql.Publisher
				gorm = gorm.Unscoped().FirstOrCreate(&publisher, sql.Publisher{Name: strings.TrimSpace(v)})
				if gorm.Error != nil {
					return offers, gorm.Error
				}
				publisherIDs = append(publisherIDs, publisher.ID)
			}

			b, err = json.Marshal(publisherIDs)
			if err != nil {
				return offers, err
			}
			app.Publishers = string(b)

			// Developers
			gorm, err = sql.GetMySQLClient()
			if err != nil {
				return offers, err
			}

			var developerIDs []int
			for _, v := range response.Data.Developers {
				var developer sql.Developer
				gorm = gorm.Unscoped().FirstOrCreate(&developer, sql.Developer{Name: strings.TrimSpace(v)})
				if gorm.Error != nil {
					return offers, gorm.Error
				}
				developerIDs = append(developerIDs, developer.ID)
			}

			b, err = json.Marshal(developerIDs)
			if err != nil {
				return offers, err
			}
			app.Developers = string(b)

			// Categories
			var categories []int8
			for _, v := range response.Data.Categories {
				categories = append(categories, v.ID)
			}

			b, err = json.Marshal(categories)
			if err != nil {
				return offers, err
			}

			app.Categories = string(b)

			// Genres
			gorm, err = sql.GetMySQLClient()
			if err != nil {
				return offers, err
			}

			var genreIDs []int
			for _, v := range response.Data.Genres {
				var genre sql.Genre
				gorm = gorm.Unscoped().Assign(sql.Genre{Name: strings.TrimSpace(v.Description)}).FirstOrCreate(&genre, sql.Genre{ID: int(v.ID)})
				if gorm.Error != nil {
					return offers, gorm.Error
				}
				genreIDs = append(genreIDs, genre.ID)
			}

			b, err = json.Marshal(genreIDs)
			if err != nil {
				return offers, err
			}
			app.Genres = string(b)

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
			b, err = json.Marshal(platforms)
			if err != nil {
				return offers, err
			}

			app.Platforms = string(b)

			// Demos
			var demos []int
			for _, v := range response.Data.Demos {
				demos = append(demos, int(v.AppID))
			}

			b, err = json.Marshal(demos)
			if err != nil {
				return offers, err
			}
			app.DemoIDs = string(b)

			// Images
			var wg sync.WaitGroup

			// Save background image
			wg.Add(1)
			go func() {

				defer wg.Done()

				app.Background = ""

				common := app.GetCommon()
				if assets, ok := common["library_assets"]; ok {

					assetMap := map[string]interface{}{}
					err := json.Unmarshal([]byte(assets), &assetMap)
					if err != nil {
						log.Err(err)
						return
					}

					if _, ok := assetMap["library_hero"]; ok {

						bg := "https://steamcdn-a.akamaihd.net/steam/fpo_apps/" + strconv.Itoa(app.ID) + "/library_hero.jpg"
						if helpers.GetResponseCode(bg) == 200 {
							app.Background = bg
						}
					}
				}
			}()

			wg.Wait()

			// Other
			if app.Name == "" && response.Data.Name != "" {
				app.Name = strings.TrimSpace(response.Data.Name)
			}

			app.Type = response.Data.Type
			app.IsFree = response.Data.IsFree
			app.ShortDescription = response.Data.ShortDescription
			app.MetacriticScore = response.Data.Metacritic.Score
			app.MetacriticURL = response.Data.Metacritic.URL
			app.GameID = int(response.Data.Fullgame.AppID)
			app.GameName = response.Data.Fullgame.Name
			app.ReleaseDate = response.Data.ReleaseDate.Date
			app.ReleaseDateUnix = helpers.GetReleaseDateUnix(response.Data.ReleaseDate.Date)
			app.ComingSoon = response.Data.ReleaseDate.ComingSoon

			// Save offers
			for _, v := range response.Data.PackageGroups {
				for kk, vv := range v.Subs {

					prices, err := app.GetPrices()
					log.Err(err)

					offers = append(offers, mongo.Offer{
						SubID:          vv.PackageID,
						SubOrder:       kk + 1,
						AppID:          app.ID,
						AppRating:      app.ReviewsScore,
						AppReleaseDate: time.Unix(app.ReleaseDateUnix, 0),
						AppPrices:      prices.Map(),
					})
				}
			}
		}
	}

	b, err := json.Marshal(prices)
	if err != nil {
		return offers, err
	}

	app.Prices = string(b)

	return offers, nil
}

func updateAppAchievements(app *sql.App, schema steam.SchemaForGame) error {

	app.AchievementsCount = len(schema.AvailableGameStats.Achievements)

	//
	resp, b, err := helpers.GetSteam().GetGlobalAchievementPercentagesForApp(app.ID)
	err = helpers.AllowSteamCodes(err, b, []int{403, 500})
	if err != nil {
		return err
	}

	var achievementsMap = make(map[string]float64)
	for _, v := range resp.GlobalAchievementPercentage {
		achievementsMap[v.Name] = v.Percent
	}

	// Make template struct
	var total float64
	var achievements []sql.AppAchievement
	for _, v := range schema.AvailableGameStats.Achievements {
		total += achievementsMap[v.Name]
		achievements = append(achievements, sql.AppAchievement{
			Name:        v.DisplayName,
			Icon:        v.Icon,
			Description: v.Description,
			Completed:   helpers.RoundFloatTo2DP(achievementsMap[v.Name]),
			Active:      true,
		})

		delete(achievementsMap, v.Name)
	}

	if len(schema.AvailableGameStats.Achievements) == 0 {
		app.AchievementsAverageCompletion = 0
	} else {
		app.AchievementsAverageCompletion = total / float64(len(schema.AvailableGameStats.Achievements))
	}

	// Add achievements that are in global but missing in schema
	for k, v := range achievementsMap {
		achievements = append(achievements, sql.AppAchievement{
			Name:      k,
			Completed: helpers.RoundFloatTo2DP(v),
		})
	}

	b, err = json.Marshal(achievements)
	if err != nil {
		return err
	}

	app.Achievements = string(b)

	return nil
}

func updateAppSchema(app *sql.App) (schema steam.SchemaForGame, err error) {

	resp, b, err := helpers.GetSteam().GetSchemaForGame(app.ID)
	err = helpers.AllowSteamCodes(err, b, []int{400, 403})
	if err != nil {
		return schema, err
	}

	var stats []sql.AppStat
	for _, v := range resp.AvailableGameStats.Stats {
		stats = append(stats, sql.AppStat{
			Name:        v.Name,
			Default:     v.DefaultValue,
			DisplayName: v.DisplayName,
		})
	}

	b, err = json.Marshal(stats)
	if err != nil {
		return schema, err
	}

	app.Stats = string(b)
	app.Version = resp.Version

	return resp, nil
}

func updateAppItems(app *sql.App) (archive []steam.ItemDefArchive, err error) {

	meta, _, err := helpers.GetSteam().GetItemDefMeta(app.ID)
	if err != nil {
		return archive, err
	}

	if meta.Response.Digest != "" && meta.Response.Digest != app.ItemsDigest {

		archive, _, err = helpers.GetSteam().GetItemDefArchive(app.ID, meta.Response.Digest)
		if err != nil {
			return archive, err
		}

		app.Items = len(archive)
		app.ItemsDigest = meta.Response.Digest
	}

	return archive, nil
}

func updateAppNews(app *sql.App) error {

	resp, b, err := helpers.GetSteam().GetNews(app.ID, 10000)
	err = helpers.AllowSteamCodes(err, b, []int{403})
	if err != nil {
		return err
	}

	ids, err := app.GetNewsIDs()
	if err != nil {
		return err
	}

	var documents []mongo.Document
	for _, v := range resp.Items {

		if strings.TrimSpace(v.Contents) == "" {
			continue
		}

		if helpers.SliceHasInt64(ids, int64(v.GID)) {
			continue
		}

		news := mongo.Article{}
		news.ID = int64(v.GID)
		news.Title = v.Title
		news.URL = v.URL
		news.IsExternal = v.IsExternalURL
		news.Author = v.Author
		news.Contents = v.Contents
		news.FeedLabel = v.Feedlabel
		news.Date = time.Unix(v.Date, 0)
		news.FeedName = v.Feedname
		news.FeedType = int8(v.FeedType)

		news.AppID = v.AppID
		news.AppName = app.GetName()
		news.AppIcon = app.GetIcon()

		documents = append(documents, news)
		ids = append(ids, int64(v.GID))
	}

	_, err = mongo.InsertDocuments(mongo.CollectionAppArticles, documents)
	if err != nil {
		return err
	}

	// Update app column
	b, err = json.Marshal(helpers.Unique64(ids))
	if err != nil {
		return err
	}

	app.NewsIDs = string(b)
	return nil
}

func updateAppReviews(app *sql.App) error {

	resp, b, err := helpers.GetSteam().GetReviews(app.ID)
	err = helpers.AllowSteamCodes(err, b, nil)
	if err != nil {
		return err
	}

	//
	reviews := sql.AppReviewSummary{}
	reviews.Positive = resp.QuerySummary.TotalPositive
	reviews.Negative = resp.QuerySummary.TotalNegative

	// Make slice of playerIDs
	var playersSlice []int64
	for _, v := range resp.Reviews {
		playersSlice = append(playersSlice, int64(v.Author.SteamID))
	}

	// Get players
	players, err := mongo.GetPlayersByID(playersSlice, mongo.M{"_id": 1, "persona_name": 1})
	if err != nil {
		return err
	}

	// Make map of players
	var playersMap = map[int64]mongo.Player{}
	for _, player := range players {
		playersMap[player.ID] = player
	}

	// Make template slice
	for _, v := range resp.Reviews {

		var player mongo.Player
		if val, ok := playersMap[int64(v.Author.SteamID)]; ok {
			player = val
		} else {
			player = mongo.Player{}
			player.ID = int64(v.Author.SteamID)
			player.PersonaName = "Unknown"
		}

		// Remove extra new lines
		regex := regexp.MustCompile("[\n]{3,}") // After comma
		v.Review = regex.ReplaceAllString(v.Review, "\n\n")

		reviews.Reviews = append(reviews.Reviews, sql.AppReview{
			Review:     helpers.BBCodeCompiler.Compile(v.Review),
			PlayerPath: player.GetPath(),
			PlayerName: player.PersonaName,
			Created:    time.Unix(v.TimestampCreated, 0).Format(helpers.DateYear),
			VotesGood:  v.VotesUp,
			VotesFunny: v.VotesFunny,
			Vote:       v.VotedUp,
		})
	}

	// Set score
	if reviews.Positive == 0 && reviews.Negative == 0 {

		app.ReviewsScore = 0

	} else {

		// https://planspace.org/2014/08/17/how-to-sort-by-average-rating/
		var a = 1
		var b = 2
		app.ReviewsScore = (float64(reviews.Positive+a) / float64(reviews.Positive+reviews.Negative+b)) * 100
	}

	// Sort by upvotes
	sort.Slice(reviews.Reviews, func(i, j int) bool {
		return reviews.Reviews[i].VotesGood > reviews.Reviews[j].VotesGood
	})

	// Save to App
	b, err = json.Marshal(reviews)
	if err != nil {
		return err
	}

	app.Reviews = string(b)
	return nil
}

func updateAppSteamSpy(app *sql.App) error {

	query := url.Values{}
	query.Set("request", "appdetails")
	query.Set("appid", strconv.Itoa(app.ID))

	// Create request
	client := &http.Client{}
	client.Timeout = time.Second * 5

	ssURL := "https://steamspy.com/api.php?" + query.Encode()
	req, err := http.NewRequest("GET", ssURL, nil)
	if err != nil {
		return err
	}

	var response *http.Response

	// Retrying as this call can fail
	operation := func() (err error) {

		response, err = client.Do(req)
		return err
	}

	policy := backoff.NewExponentialBackOff()
	policy.InitialInterval = time.Second * 1
	policy.MaxElapsedTime = time.Second * 10

	err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { logInfo(err) })
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return errors.New("steamspy is down: " + ssURL)
	}

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	err = response.Body.Close()
	logError(err)

	if strings.Contains(string(bytes), "Connection failed") {
		return errors.New("steamspy is down: " + ssURL)
	}

	// Unmarshal JSON
	resp := sql.SteamSpyAppResponse{}
	err = helpers.Unmarshal(bytes, &resp)
	if err != nil {
		return errors.New("steamspy is down: " + ssURL)
	}

	owners := resp.GetOwners()

	ss := sql.AppSteamSpy{
		SSAveragePlaytimeTwoWeeks: resp.Average2Weeks,
		SSAveragePlaytimeForever:  resp.AverageForever,
		SSMedianPlaytimeTwoWeeks:  resp.Median2Weeks,
		SSMedianPlaytimeForever:   resp.MedianForever,
		SSOwnersLow:               owners[0],
		SSOwnersHigh:              owners[1],
	}

	b, err := json.Marshal(ss)
	if err != nil {
		return err
	}

	app.SteamSpy = string(b)

	return nil
}

//noinspection RegExpRedundantEscape
var bundlesRegex = regexp.MustCompile(`store\.steampowered\.com\/app\/[0-9]+$`)

func updateBundlesAndOffers(app *sql.App) error {

	// This app causes infinite redirects..
	if app.ID == 12820 {
		return nil
	}

	// Skip these app types
	if helpers.SliceHasString([]string{"media", "movie"}, app.Type) {
		return nil
	}

	var bundleIDs []string

	// Retry call
	operation := func() (err error) {

		bundleIDs = []string{}

		c := colly.NewCollector(
			colly.URLFilters(bundlesRegex),
		)

		// Bundles
		c.OnHTML("div.game_area_purchase_game_wrapper input[name=bundleid]", func(e *colly.HTMLElement) {
			bundleIDs = append(bundleIDs, e.Attr("value"))
		})

		// // Offers
		// var offers []mongo.Offer
		//
		// c.OnHTML("div.game_area_purchase_game_wrapper", func(e *colly.HTMLElement) {
		//
		// 	offer := mongo.Offer{}
		//
		// 	e.ForEach("[name=subid]", func(_ int, el *colly.HTMLElement) {
		//
		// 		i, err := strconv.Atoi(el.Attr("value"))
		// 		if err != nil {
		// 			log.Err(err)
		// 		} else {
		// 			offer.SubID = i
		// 		}
		// 	})
		//
		// 	e.ForEach("game_purchase_discount_countdown", func(_ int, el *colly.HTMLElement) {
		//
		// 	})
		//
		// 	offers = append(offers, offer)
		// })

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

	err := backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { logInfo(err, app.ID) })
	if err != nil {
		return err
	}

	//
	var IDInts = helpers.StringSliceToIntSlice(bundleIDs)

	for _, v := range IDInts {
		err := ProduceBundle(v, app.ID)
		if err != nil {
			return err
		}
	}

	b, err := json.Marshal(IDInts)
	if err != nil {
		return err
	}

	app.BundleIDs = string(b)

	return nil
}

func saveAppToInflux(app sql.App) (err error) {

	reviews, err := app.GetReviews()
	if err != nil {
		return err
	}

	fields := map[string]interface{}{
		"reviews_score":    app.ReviewsScore,
		"reviews_positive": reviews.Positive,
		"reviews_negative": reviews.Negative,
	}

	price := app.GetPrice(steam.ProductCCUS)
	if price.Exists {
		fields["price_us_initial"] = price.Initial
		fields["price_us_final"] = price.Final
		fields["price_us_discount"] = price.DiscountPercent
	}

	_, err = helpers.InfluxWrite(helpers.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(helpers.InfluxMeasurementApps),
		Tags: map[string]string{
			"app_id": strconv.Itoa(app.ID),
		},
		Fields:    fields,
		Time:      time.Now(),
		Precision: "m",
	})

	return err
}

func updateAppTwitch(app *sql.App) error {

	client, err := helpers.GetTwitch()
	if err != nil {
		return err
	}

	if app.TwitchID == 0 && app.Name != "" {

		var resp *helix.GamesResponse

		// Retrying as this call can fail
		operation := func() (err error) {

			resp, err = client.GetGames(&helix.GamesParams{Names: []string{app.Name}})
			return err
		}

		policy := backoff.NewExponentialBackOff()

		err = backoff.RetryNotify(operation, backoff.WithMaxRetries(policy, 3), func(err error, t time.Duration) { logInfo(err) })
		if err != nil {
			return err
		}

		if len(resp.Data.Games) > 0 {

			i, err := strconv.Atoi(resp.Data.Games[0].ID)
			if err != nil {
				return err
			}

			app.TwitchID = i
		}
	}

	return nil
}

func getCurrentAppItems(appID int) (items []int, err error) {

	resp, err := mongo.GetAppItems(0, 0, mongo.M{"app_id": appID}, mongo.M{"item_def_id": 1})
	for _, v := range resp {
		items = append(items, v.ItemDefID)
	}

	return items, err
}

func updateAppPlaytimeStats(app *sql.App) (err error) {

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

func saveOffers(app sql.App, offers []mongo.Offer) (err error) {

	return nil
}

func saveAppItems(appID int, newItems []steam.ItemDefArchive, currentItemIDs []int) (err error) {

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
				ItemQuality:      v.ItemQuality,
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

	// Find removed items
	var oldDocumentIDs []int
	for _, v := range currentItemIDs {
		_, ok := newItemsMap[v]
		if !ok {
			oldDocumentIDs = append(oldDocumentIDs, v)
		}
	}
	err = mongo.DeleteAppItems(appID, oldDocumentIDs)

	return nil
}
