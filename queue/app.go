package queue

import (
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/cenkalti/backoff"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/websockets"
	"github.com/gocolly/colly"
	"github.com/streadway/amqp"
)

type RabbitMessageApp struct {
	PICSAppInfo RabbitMessageProduct
	Payload     produceAppPayload
}

func (d RabbitMessageApp) getConsumeQueue() RabbitQueue {
	return QueueAppsData
}

func (d RabbitMessageApp) getProduceQueue() RabbitQueue {
	return QueueApps
}

func (d RabbitMessageApp) getRetryData() RabbitMessageDelay {
	return RabbitMessageDelay{}
}

func (d RabbitMessageApp) process(msg amqp.Delivery) (requeue bool, err error) {

	// Get message payload
	rabbitMessage := RabbitMessageApp{}

	err = helpers.Unmarshal(msg.Body, &rabbitMessage)
	if err != nil {
		return false, err
	}

	message := rabbitMessage.PICSAppInfo

	logInfo("Consuming app: " + strconv.Itoa(message.ID))

	if !db.IsValidAppID(message.ID) {
		return false, errors.New("invalid app ID: " + strconv.Itoa(message.ID))
	}

	// Load current app
	gorm, err := db.GetMySQLClient()
	if err != nil {
		return true, err
	}

	app := db.App{}
	gorm = gorm.FirstOrInit(&app, db.App{ID: message.ID})
	if gorm.Error != nil {
		return true, gorm.Error
	}

	// Skip if updated in last day, unless its from PICS
	if app.UpdatedAt.Unix() > time.Now().Add(time.Hour * -24).Unix() && app.ChangeNumber >= message.ChangeNumber {
		logInfo("Skipping, updated in last day")
		return false, nil
	}

	var appBeforeUpdate = app

	err = updateAppPICS(&app, rabbitMessage)
	if err != nil {
		return true, err
	}

	err = updateAppDetails(&app)
	if err != nil && err != steam.ErrAppNotFound {
		return true, err
	}

	schema, err := updateAppSchema(&app)
	if err != nil {
		return true, err
	}

	err = updateAppAchievements(&app, schema)
	if err != nil {
		return true, err
	}

	err = updateAppNews(&app)
	if err != nil {
		return true, err
	}

	err = updateAppReviews(&app)
	if err != nil {
		return true, err
	}

	err = updateAppSteamSpy(&app)
	if err != nil {
		return true, err
	}

	err = updateBundles(&app)
	if err != nil {
		return true, err
	}

	// Save price changes
	err = savePriceChanges(appBeforeUpdate, app)
	if err != nil {
		return true, err
	}

	// Misc
	app.Type = strings.ToLower(app.Type)
	app.ReleaseState = strings.ToLower(app.ReleaseState)

	// Save new data
	gorm = gorm.Save(&app)
	if gorm.Error != nil {
		return true, gorm.Error
	}

	// Send websocket
	page, err := websockets.GetPage(websockets.PageApp)
	if err != nil {
		return true, err
	} else if page.HasConnections() {
		page.Send(app.ID)
	}

	return false, nil
}

func updateAppPICS(app *db.App, rabbitMessage RabbitMessageApp) (err error) {

	message := rabbitMessage.PICSAppInfo

	if message.ChangeNumber > app.ChangeNumber {
		app.ChangeNumberDate = time.Unix(rabbitMessage.Payload.Time, 0)
	}

	app.ID = message.ID
	app.ChangeNumber = message.ChangeNumber

	for _, v := range message.KeyValues.Children {

		switch v.Name {
		case "appid":

			// No need for this
			//var i64 int64
			//i64, err = strconv.ParseInt(v.Value.(string), 10, 32)
			//if err != nil {
			//	return err
			//}
			//app.ID = int(i64)

		case "common":

			var common = db.PICSAppCommon{}
			for _, vv := range v.Children {
				if vv.Value == nil {
					b, err := json.Marshal(vv.ToNestedMaps())
					if err != nil {
						return err
					}
					common[vv.Name] = string(b)
				} else {
					common[vv.Name] = vv.Value.(string)
				}
			}
			err = app.SetCommon(common)
			if err != nil {
				return err
			}

		case "extended":

			err = app.SetExtended(v.GetExtended())
			if err != nil {
				return err
			}

		case "config":

			config, launch := v.GetAppConfig()

			err = app.SetConfig(config)
			if err != nil {
				return err
			}

			err = app.SetLaunch(launch)
			if err != nil {
				return err
			}

		case "depots":

			err = app.SetDepots(v.GetAppDepots())
			if err != nil {
				return err
			}

		case "public_only":

			if v.Value.(string) == "1" {
				app.PublicOnly = true
			}

		case "ufs":

			var common = db.PICSAppUFS{}
			for _, vv := range v.Children {
				if vv.Value == nil {
					b, err := json.Marshal(vv.ToNestedMaps())
					if err != nil {
						return err
					}
					common[vv.Name] = string(b)
				} else {
					common[vv.Name] = vv.Value.(string)
				}
			}
			err = app.SetUFS(common)
			if err != nil {
				return err
			}

		case "install":

			err = app.SetInstall(v.ToNestedMaps())
			if err != nil {
				return err
			}

		case "localization":

			err = app.SetLocalization(v.ToNestedMaps())
			if err != nil {
				return err
			}

		case "sysreqs":

			err = app.SetSystemRequirements(v.ToNestedMaps())
			if err != nil {
				return err
			}

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
			b, err = json.Marshal(response.Data.Publishers)
			if err != nil {
				return err
			}

			app.Publishers = string(b)

			// Developers
			b, err = json.Marshal(response.Data.Developers)
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
			b, err = json.Marshal(response.Data.Genres)
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

			// Other
			app.Name = response.Data.Name
			app.Type = response.Data.Type
			app.IsFree = response.Data.IsFree
			app.ShortDescription = response.Data.ShortDescription
			app.HeaderImage = response.Data.HeaderImage
			app.MetacriticScore = response.Data.Metacritic.Score
			app.MetacriticURL = response.Data.Metacritic.URL
			app.Background = response.Data.Background
			app.GameID = response.Data.Fullgame.AppID
			app.GameName = response.Data.Fullgame.Name
			app.ReleaseDate = response.Data.ReleaseDate.Date
			app.ReleaseDateUnix = helpers.GetReleaseDateUnix(response.Data.ReleaseDate.Date)
			app.ComingSoon = response.Data.ReleaseDate.ComingSoon
		}
	}

	return app.SetPrices(prices)
}

func updateAppAchievements(app *db.App, schema steam.SchemaForGame) error {

	resp, _, err := helpers.GetSteam().GetGlobalAchievementPercentagesForApp(app.ID)

	// This endpoint seems to error if the app has no achievement data, so it's probably fine.
	err2, ok := err.(steam.Error)
	if ok && (err2.Code() == 403 || err2.Code() == 500) {
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
			Name:        v.Name,
			Icon:        v.Icon,
			Description: v.Description,
			Completed:   helpers.RoundFloatTo2DP(achievementsMap[v.Name]),
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
	if ok && (err2.Code() == 403) {
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
	if ok && (err2.Code() == 403) {
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

		if helpers.SliceHasInt64(ids, v.GID) {
			continue
		}

		kinds = append(kinds, db.CreateArticle(*app, v))
	}

	err = db.BulkSaveKinds(kinds, db.KindNews, false)
	if err != nil {
		return err
	}

	err = app.SetNewsIDs(resp)
	if err != nil {
		return err
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
		playersSlice = append(playersSlice, v.Author.SteamID)
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
		if val, ok := playersMap[v.Author.SteamID]; ok {
			player = val
		} else {
			player = db.Player{}
			player.PlayerID = v.Author.SteamID
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

	// Save to App
	b, err := json.Marshal(reviews)
	if err != nil {
		return err
	}

	app.Reviews = string(b)

	// Log this app score
	err = db.SaveAppOverTime(*app, reviews)
	if err != nil {
		return err
	}

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

	err = backoff.Retry(operation, policy)
	if err != nil {
		return err
	}

	defer func(body io.ReadCloser) {
		if body != nil {
			err = body.Close()
			log.Err(err)
		}
	}(response.Body)

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

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
	)

	c.OnHTML("div.game_area_purchase_game_wrapper input[name=bundleid]", func(e *colly.HTMLElement) {
		bundleIDs = append(bundleIDs, e.Attr("value"))
	})

	err := c.Visit("https://store.steampowered.com/app/" + strconv.Itoa(app.ID))
	if err != nil {
		if !strings.Contains(err.Error(), "because its not in AllowedDomains") {
			return err
		}
	}

	var IDInts = helpers.StringSliceToIntSlice(bundleIDs)

	for _, v := range IDInts {
		err := QueueBundle(v)
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
