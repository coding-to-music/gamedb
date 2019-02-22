package queue

import (
	"encoding/json"
	"errors"
	"html/template"
	"io/ioutil"
	"math"
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
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/websockets"
	"github.com/gocolly/colly"
	influx "github.com/influxdata/influxdb1-client"
	"github.com/mitchellh/mapstructure"
	"github.com/streadway/amqp"
)

type appMessage struct {
	ID          int                  `json:"id"`
	PICSAppInfo rabbitMessageProduct `` // Leave JSON key name as default
}

type appQueue struct {
	baseQueue
}

func (q appQueue) processMessages(msgs []amqp.Delivery) {

	msg := msgs[0]

	var err error
	var payload = baseMessage{
		Message: appMessage{},
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

	if !db.IsValidAppID(message.ID) {
		logError(errors.New("invalid app ID: " + strconv.Itoa(message.ID)))
		payload.ack(msg)
		return
	}

	// Load current app
	gorm, err := db.GetMySQLClient()
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	app := db.App{}
	gorm = gorm.FirstOrInit(&app, db.App{ID: message.ID})
	if gorm.Error != nil {
		logError(gorm.Error, message.ID)
		payload.ackRetry(msg)
		return
	}

	// Skip if updated in last day, unless its from PICS
	if app.UpdatedAt.Unix() > time.Now().Add(time.Hour * -24).Unix() {
		if app.ChangeNumber >= message.PICSAppInfo.ChangeNumber {
			logInfo("Skipping app, updated in last day")
			payload.ack(msg)
			return
		}
	}

	var appBeforeUpdate = app

	if message.PICSAppInfo.ID > 0 {
		err = updateAppPICS(&app, payload, message)
		if err != nil {
			logError(err, message.ID)
			payload.ackRetry(msg)
			return
		}
	}

	err = updateAppDetails(&app)
	if err != nil && err != steam.ErrAppNotFound {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	schema, err := updateAppSchema(&app)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	err = updateAppAchievements(&app, schema)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	err = updateAppNews(&app)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	err = updateAppPlayerCount(&app)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	err = updateAppReviews(&app)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	err = updateAppSteamSpy(&app)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	err = updateBundles(&app)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	// Save price changes
	err = savePriceChanges(appBeforeUpdate, app)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	// Misc
	app.Type = strings.ToLower(app.Type)
	app.ReleaseState = strings.ToLower(app.ReleaseState)

	// Save new data
	gorm = gorm.Save(&app)
	if gorm.Error != nil {
		logError(gorm.Error, message.ID)
		payload.ackRetry(msg)
		return
	}

	// Save to InfluxDB
	err = saveAppToInflux(app)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	// Send websocket
	page, err := websockets.GetPage(websockets.PageApp)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	if page.HasConnections() {
		page.Send(message.ID)
	}

	payload.ack(msg)
}

func updateAppPICS(app *db.App, payload baseMessage, message appMessage) (err error) {

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

			var common = db.PICSAppCommon{}
			var tags []int

			for _, vv := range v.Children {

				if vv.Name == "store_tags" {
					stringTags := vv.GetChildrenAsSlice()
					tags = helpers.StringSliceToIntSlice(stringTags)
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

			var ufs = db.PICSAppUFS{}
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

			b, err := json.Marshal(v.ToNestedMaps())
			if err != nil {
				return err
			}

			app.Localization = string(b)

		case "sysreqs":

			b, err := json.Marshal(v.ToNestedMaps())
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

func updateAppDetails(app *db.App) error {

	prices := db.ProductPrices{}

	for _, code := range helpers.GetActiveCountries() {

		// Get app details
		response, _, err := helpers.GetSteam().GetAppDetails(app.ID, code, steam.LanguageEnglish)
		if err != nil && err != steam.ErrAppNotFound {
			return err
		}

		prices.AddPriceFromApp(code, response)

		if code == steam.CountryUS {

			// Screenshots
			var images []db.AppImage
			for _, v := range response.Data.Screenshots {
				images = append(images, db.AppImage{
					PathFull:      v.PathFull,
					PathThumbnail: v.PathThumbnail,
				})
			}

			b, err := json.Marshal(images)
			if err != nil {
				return err
			}

			app.Screenshots = string(b)

			// Movies
			var videos []db.AppVideo
			for _, v := range response.Data.Movies {
				videos = append(videos, db.AppVideo{
					PathFull:      v.Webm.Max,
					PathThumbnail: v.Thumbnail,
					Title:         v.Name,
				})
			}

			b, err = json.Marshal(videos)
			if err != nil {
				return err
			}

			app.Movies = string(b)

			// DLC
			b, err = json.Marshal(response.Data.DLC)
			if err != nil {
				return err
			}

			app.DLC = string(b)
			app.DLCCount = len(response.Data.DLC)

			// Packages
			b, err = json.Marshal(response.Data.Packages)
			if err != nil {
				return err
			}

			app.Packages = string(b)

			// Publishers
			gorm, err := db.GetMySQLClient()
			if err != nil {
				return err
			}

			var publisherIDs []int
			for _, v := range response.Data.Publishers {
				var publisher db.Publisher
				gorm = gorm.Unscoped().FirstOrCreate(&publisher, db.Publisher{Name: strings.TrimSpace(v)})
				if gorm.Error != nil {
					return gorm.Error
				}
				publisherIDs = append(publisherIDs, publisher.ID)
			}

			b, err = json.Marshal(publisherIDs)
			if err != nil {
				return err
			}
			app.Publishers = string(b)

			// Developers
			gorm, err = db.GetMySQLClient()
			if err != nil {
				return err
			}

			var developerIDs []int
			for _, v := range response.Data.Developers {
				var developer db.Developer
				gorm = gorm.Unscoped().FirstOrCreate(&developer, db.Developer{Name: strings.TrimSpace(v)})
				if gorm.Error != nil {
					return gorm.Error
				}
				developerIDs = append(developerIDs, developer.ID)
			}

			b, err = json.Marshal(developerIDs)
			if err != nil {
				return err
			}
			app.Developers = string(b)

			// Categories
			var categories []int8
			for _, v := range response.Data.Categories {
				categories = append(categories, v.ID)
			}

			b, err = json.Marshal(categories)
			if err != nil {
				return err
			}

			app.Categories = string(b)

			// Genres
			gorm, err = db.GetMySQLClient()
			if err != nil {
				return err
			}

			var genreIDs []int
			for _, v := range response.Data.Genres {
				var genre db.Genre
				gorm = gorm.Unscoped().Assign(db.Genre{Name: strings.TrimSpace(v.Description)}).FirstOrCreate(&genre, db.Genre{ID: int(v.ID)})
				if gorm.Error != nil {
					return gorm.Error
				}
				genreIDs = append(genreIDs, genre.ID)
			}

			b, err = json.Marshal(genreIDs)
			if err != nil {
				return err
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
				return err
			}

			app.Platforms = string(b)

			// Demos
			var demos []int
			for _, v := range response.Data.Demos {
				demos = append(demos, int(v.AppID))
			}

			b, err = json.Marshal(demos)
			if err != nil {
				return err
			}
			app.DemoIDs = string(b)

			// Images
			var wg sync.WaitGroup

			wg.Add(1)
			go func() {

				defer wg.Done()

				code := helpers.GetResponseCode(response.Data.Background)
				app.Background = ""
				if code == 200 {
					app.Background = response.Data.Background
				}
			}()

			wg.Add(1)
			go func() {

				defer wg.Done()

				code := helpers.GetResponseCode(response.Data.HeaderImage)
				app.HeaderImage = ""
				if code == 200 {
					app.HeaderImage = response.Data.HeaderImage
				}
			}()

			wg.Wait()

			// Other
			app.Name = response.Data.Name
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
		}
	}

	b, err := json.Marshal(prices)
	if err != nil {
		return err
	}

	app.Prices = string(b)

	return nil
}

func updateAppAchievements(app *db.App, schema steam.SchemaForGame) error {

	resp, _, err := helpers.GetSteam().GetGlobalAchievementPercentagesForApp(app.ID)

	// This endpoint seems to error if the app has no achievement data, so it's probably fine.
	err2, ok := err.(steam.Error)
	if ok && (err2.Code == 403 || err2.Code == 500) {
		return nil
	}
	if err != nil {
		return err
	}

	var achievementsMap = make(map[string]float64)
	for _, v := range resp.GlobalAchievementPercentage {
		achievementsMap[v.Name] = v.Percent
	}

	// Make template struct
	var achievements []db.AppAchievement
	for _, v := range schema.AvailableGameStats.Achievements {
		achievements = append(achievements, db.AppAchievement{
			Name:        v.DisplayName,
			Icon:        v.Icon,
			Description: v.Description,
			Completed:   helpers.RoundFloatTo2DP(achievementsMap[v.Name]),
		})

		delete(achievementsMap, v.Name)
	}

	// Add achievements that are in global but missing in schema
	for k, v := range achievementsMap {
		achievements = append(achievements, db.AppAchievement{
			Name:      k,
			Completed: helpers.RoundFloatTo2DP(v),
		})
	}

	b, err := json.Marshal(achievements)
	if err != nil {
		return err
	}

	app.Achievements = string(b)

	return nil
}

func updateAppSchema(app *db.App) (schema steam.SchemaForGame, err error) {

	resp, _, err := helpers.GetSteam().GetSchemaForGame(app.ID)

	// This endpoint seems to error if the app has no schema, so it's probably fine.
	err2, ok := err.(steam.Error)
	if ok && (err2.Code == 403 || err2.Code == 400) {
		return schema, nil
	}
	if err != nil {
		return schema, err
	}

	var stats []db.AppStat
	for _, v := range resp.AvailableGameStats.Stats {
		stats = append(stats, db.AppStat{
			Name:        v.Name,
			Default:     v.DefaultValue,
			DisplayName: v.DisplayName,
		})
	}

	b, err := json.Marshal(stats)
	if err != nil {
		return schema, err
	}

	app.Stats = string(b)
	app.Version = resp.Version

	return resp, nil
}

func updateAppNews(app *db.App) error {

	resp, _, err := helpers.GetSteam().GetNews(app.ID, 10000)

	// This endpoint seems to error if the app has no news, so it's probably fine.
	err2, ok := err.(steam.Error)
	if ok && (err2.Code == 403) {
		return nil
	}
	if err != nil {
		return err
	}

	ids, err := app.GetNewsIDs()
	if err != nil {
		return err
	}

	var kinds []db.Kind
	for _, v := range resp.Items {

		if strings.TrimSpace(v.Contents) == "" {
			continue
		}

		if helpers.SliceHasInt64(ids, int64(v.GID)) {
			continue
		}

		news := db.News{}
		news.ArticleID = int64(v.GID)
		news.Title = v.Title
		news.URL = v.URL
		news.IsExternal = v.IsExternalURL
		news.Author = v.Author
		news.Contents = v.Contents
		news.FeedLabel = v.Feedlabel
		news.Date = time.Unix(int64(v.Date), 0)
		news.FeedName = v.Feedname
		news.FeedType = int8(v.FeedType)

		news.AppID = v.AppID
		news.AppName = app.Name
		news.AppIcon = app.Icon

		kinds = append(kinds, news)
		ids = append(ids, int64(v.GID))
	}

	err = db.BulkSaveKinds(kinds, db.KindNews, false)
	if err != nil {
		return err
	}

	// Update app column
	b, err := json.Marshal(helpers.Unique64(ids))
	if err != nil {
		return err
	}

	app.NewsIDs = string(b)
	return nil
}

func updateAppPlayerCount(app *db.App) error {

	resp, _, err := helpers.GetSteam().GetNumberOfCurrentPlayers(app.ID)

	err2, ok := err.(steam.Error)
	if ok && (err2.Code == 404) {
		err = nil
	}
	if err != nil {
		return err
	}

	if resp > app.PlayerCount {
		app.PlayerCount = resp
	}

	return nil
}

func updateAppReviews(app *db.App) error {

	resp, _, err := helpers.GetSteam().GetReviews(app.ID)
	if err != nil {
		return err
	}

	//
	reviews := db.AppReviewSummary{}
	reviews.Positive = resp.QuerySummary.TotalPositive
	reviews.Negative = resp.QuerySummary.TotalNegative

	// Make slice of playerIDs
	var playersSlice []int64
	for _, v := range resp.Reviews {
		playersSlice = append(playersSlice, int64(v.Author.SteamID))
	}

	// Get players from Datastore
	players, err := db.GetPlayersByIDs(playersSlice)
	if err != nil {
		return err
	}

	// Make map of players
	var playersMap = map[int64]db.Player{}
	for _, v := range players {
		playersMap[v.PlayerID] = v
	}

	// Make template slice
	for _, v := range resp.Reviews {

		var player db.Player
		if val, ok := playersMap[int64(v.Author.SteamID)]; ok {
			player = val
		} else {
			player = db.Player{}
			player.PlayerID = int64(v.Author.SteamID)
			player.PersonaName = "Unknown"
		}

		// Remove extra new lines
		regex := regexp.MustCompile("[\n]{3,}") // After comma
		v.Review = regex.ReplaceAllString(v.Review, "\n\n")

		reviews.Reviews = append(reviews.Reviews, db.AppReview{
			Review:     template.HTML(helpers.BBCodeCompiler.Compile(v.Review)),
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

		total := float64(reviews.Positive + reviews.Negative)
		average := float64(reviews.Positive) / total
		score := average - (average-0.5)*math.Pow(2, -math.Log10(total + 1))

		app.ReviewsScore = helpers.RoundFloatTo2DP(score * 100)
	}

	// Sort by upvotes
	sort.Slice(reviews.Reviews, func(i, j int) bool {
		return reviews.Reviews[i].VotesGood > reviews.Reviews[j].VotesGood
	})

	// Save to App
	b, err := json.Marshal(reviews)
	if err != nil {
		return err
	}

	app.Reviews = string(b)
	return nil
}

func updateAppSteamSpy(app *db.App) error {

	query := url.Values{}
	query.Set("request", "appdetails")
	query.Set("appid", strconv.Itoa(app.ID))

	// Create request
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://steamspy.com/api.php?"+query.Encode(), nil)
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
		return errors.New("steamspy is down")
	}

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	err = response.Body.Close()
	logError(err)

	// Unmarshal JSON
	resp := db.SteamSpyAppResponse{}
	err = helpers.Unmarshal(bytes, &resp)
	if err != nil {
		return err
	}

	owners := resp.GetOwners()

	ss := db.AppSteamSpy{
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

func updateBundles(app *db.App) error {

	var bundleIDs []string

	c := colly.NewCollector(
		colly.AllowedDomains("store.steampowered.com"),
		colly.AllowURLRevisit(), // This is for retrys
	)

	c.OnHTML("div.game_area_purchase_game_wrapper input[name=bundleid]", func(e *colly.HTMLElement) {
		bundleIDs = append(bundleIDs, e.Attr("value"))
	})

	// Retry call
	operation := func() (err error) {

		err = c.Visit("https://store.steampowered.com/app/" + strconv.Itoa(app.ID))
		if err != nil && !strings.Contains(err.Error(), "because its not in AllowedDomains") {
			return backoff.Permanent(err)
		}
		return err
	}

	policy := backoff.NewExponentialBackOff()

	err := backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { logInfo(err) })
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

func saveAppToInflux(app db.App) error {

	price, err := app.GetPrice(steam.CountryUS)
	if err != nil && err != db.ErrMissingCountryCode {
		return err
	}

	reviews, err := app.GetReviews()
	if err != nil && err != db.ErrMissingCountryCode {
		return err
	}

	_, err = db.InfluxWriteMany(db.InfluxRetentionPolicyAllTime, influx.BatchPoints{
		Points: []influx.Point{
			{
				Measurement: string(db.InfluxMeasurementApps),
				Tags: map[string]string{
					"app_id": strconv.Itoa(app.ID),
				},
				Fields: map[string]interface{}{
					"reviews_score":     app.ReviewsScore,
					"reviews_positive":  reviews.Positive,
					"reviews_negative":  reviews.Negative,
					"price_us_initial":  price.Initial,
					"price_us_final":    price.Final,
					"price_us_discount": price.DiscountPercent,
				},
				Time:      time.Now(),
				Precision: "h",
			},
			{
				Measurement: string(db.InfluxMeasurementApps),
				Tags: map[string]string{
					"app_id": strconv.Itoa(app.ID),
				},
				Fields: map[string]interface{}{
					"player_count": app.PlayerCount,
				},
				Time:      time.Now(),
				Precision: "m",
			},
		},
	})

	return err
}
