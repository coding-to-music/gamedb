package queue

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
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
	if app.UpdatedAt.Unix() > time.Now().Add(time.Hour * -24).Unix() && app.PICSChangeNumber >= message.ChangeNumber {
		logInfo("Skipping, updated in last day")
		//return false, nil
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

	err = updateAppAchievements(&app)
	if err != nil {
		return true, err
	}

	err = updateAppSchema(&app)
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

	if message.ChangeNumber > app.PICSChangeNumber {
		app.PICSChangeNumberDate = time.Unix(rabbitMessage.Payload.Time, 0)
	}

	app.ID = message.ID
	app.PICSChangeNumber = message.ChangeNumber

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
					bytes, err := json.Marshal(vv.ToNestedMaps())
					if err != nil {
						return err
					}
					common[vv.Name] = string(bytes)
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
				app.PICSPublicOnly = true
			}

		case "ufs":

			var common = db.PICSAppUFS{}
			for _, vv := range v.Children {
				if vv.Value == nil {
					bytes, err := json.Marshal(vv.ToNestedMaps())
					if err != nil {
						return err
					}
					common[vv.Name] = string(bytes)
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
			logWarning(v.Name + " field in app PICS ignored (Change " + strconv.Itoa(app.PICSChangeNumber) + ")")
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
			screenshotsString, err := json.Marshal(response.Data.Screenshots)
			if err != nil {
				return err
			}

			// Movies
			moviesString, err := json.Marshal(response.Data.Movies)
			if err != nil {
				return err
			}

			// Achievements
			achievementsString, err := json.Marshal(response.Data.Achievements)
			if err != nil {
				return err
			}

			// DLC
			dlcString, err := json.Marshal(response.Data.DLC)
			if err != nil {
				return err
			}

			// Packages
			packagesString, err := json.Marshal(response.Data.Packages)
			if err != nil {
				return err
			}

			// Publishers
			publishersString, err := json.Marshal(response.Data.Publishers)
			if err != nil {
				return err
			}

			// Developers
			developersString, err := json.Marshal(response.Data.Developers)
			if err != nil {
				return err
			}

			// Categories
			var categories []int8
			for _, v := range response.Data.Categories {
				categories = append(categories, v.ID)
			}

			categoriesString, err := json.Marshal(categories)
			if err != nil {
				return err
			}

			genresString, err := json.Marshal(response.Data.Genres)
			if err != nil {
				return err
			}

			// Platforms
			var platforms []string
			if response.Data.Platforms.Linux {
				platforms = append(platforms, "linux")
			}
			if response.Data.Platforms.Windows {
				platforms = append(platforms, "windows")
			}
			if response.Data.Platforms.Windows {
				platforms = append(platforms, "macos")
			}

			platformsString, err := json.Marshal(platforms)
			if err != nil {
				return err
			}

			// Other
			app.Name = response.Data.Name
			app.Type = response.Data.Type
			app.IsFree = response.Data.IsFree
			app.DLC = string(dlcString)
			app.DLCCount = len(response.Data.DLC)
			app.ShortDescription = response.Data.ShortDescription
			app.HeaderImage = response.Data.HeaderImage
			app.Developers = string(developersString)
			app.Publishers = string(publishersString)
			app.Packages = string(packagesString)
			app.MetacriticScore = response.Data.Metacritic.Score
			app.MetacriticURL = response.Data.Metacritic.URL
			app.Categories = string(categoriesString)
			app.Genres = string(genresString)
			app.Screenshots = string(screenshotsString)
			app.Movies = string(moviesString)
			app.Achievements = string(achievementsString)
			app.Background = response.Data.Background
			app.Platforms = string(platformsString)
			app.GameID = response.Data.Fullgame.AppID
			app.GameName = response.Data.Fullgame.Name
			app.ReleaseDate = response.Data.ReleaseDate.Date
			app.ReleaseDateUnix = helpers.GetReleaseDateUnix(response.Data.ReleaseDate.Date)
			app.ComingSoon = response.Data.ReleaseDate.ComingSoon
		}
	}

	return app.SetPrices(prices)
}

func updateAppAchievements(app *db.App) error {

	percentages, _, err := helpers.GetSteam().GetGlobalAchievementPercentagesForApp(app.ID)

	// This endpoint seems to error if the app has no achievement data, so it's probably fine.
	err2, ok := err.(steam.Error)
	if ok && (err2.Code() == 403 || err2.Code() == 500) {
		return nil
	}
	if err != nil {
		return err
	}

	percentagesBytes, err := json.Marshal(percentages)
	if err != nil {
		return err
	}

	app.AchievementPercentages = string(percentagesBytes)

	return nil
}

func updateAppSchema(app *db.App) error {

	schema, _, err := helpers.GetSteam().GetSchemaForGame(app.ID)

	// This endpoint seems to error if the app has no schema, so it's probably fine.
	err2, ok := err.(steam.Error)
	if ok && (err2.Code() == 403) {
		return nil
	}
	if err != nil {
		return err
	}

	schemaString, err := json.Marshal(schema)
	if err != nil {
		return err
	}

	app.Schema = string(schemaString)

	return nil
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

	var kinds []db.Kind
	for _, v := range resp.Items {

		if strings.TrimSpace(v.Contents) == "" {
			continue
		}

		ids, err := app.GetNewsIDs()
		if err != nil {
			return err
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

	var reviewsResp steam.ReviewsResponse

	reviewsResp, _, err := helpers.GetSteam().GetReviews(app.ID)
	if err != nil {
		return err
	}

	reviewsBytes, err := json.Marshal(reviewsResp)
	if err != nil {
		return err
	}

	app.Reviews = string(reviewsBytes)
	app.ReviewsPositive = reviewsResp.QuerySummary.TotalPositive
	app.ReviewsNegative = reviewsResp.QuerySummary.TotalNegative
	app.SetReviewScore()

	// Log this app score
	err = db.SaveAppOverTime(*app)
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
	resp := db.SteamSpyApp{}
	err = helpers.Unmarshal(bytes, &resp)
	if err != nil {
		return err
	}

	app.SSAveragePlaytimeForever = resp.AverageForever
	app.SSAveragePlaytimeTwoWeeks = resp.Average2Weeks
	app.SSMedianPlaytimeForever = resp.MedianForever
	app.SSMedianPlaytimeTwoWeeks = resp.Median2Weeks

	owners := resp.GetOwners()
	app.SSOwnersLow = owners[0]
	app.SSOwnersHigh = owners[1]

	return nil
}

func updateBundles(app *db.App) error {

	var IDStrings []string

	c := colly.NewCollector(
		colly.AllowedDomains("store.steampowered.com"),
	)

	c.OnHTML("div.game_area_purchase_game_wrapper input[name=bundleid]", func(e *colly.HTMLElement) {
		IDStrings = append(IDStrings, e.Attr("value"))
	})

	err := c.Visit("https://store.steampowered.com/app/" + strconv.Itoa(app.ID))
	if err != nil {
		return err
	}

	var IDInts = helpers.StringSliceToIntSlice(IDStrings)

	b, err := json.Marshal(IDInts)
	if err != nil {
		return err
	}

	app.BundleIDs = string(b)

	return nil
}
