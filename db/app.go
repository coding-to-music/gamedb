package db

import (
	"encoding/json"
	"errors"
	"html/template"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/logging"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
)

const (
	platformWindows = "windows"
	platformMac     = "macos"
	platformLinux   = "linux"

	DefaultAppIcon = "/assets/img/no-app-image-square.jpg"
)

var (
	ErrInvalidAppID = UpdateError{"invalid app id", true, false}
	ErrCantFindApp  = errors.New("no such app")
)

type App struct {
	ID                     int        `gorm:"not null;column:id;primary_key"`
	CreatedAt              *time.Time `gorm:"not null;column:created_at"`
	UpdatedAt              *time.Time `gorm:"not null;column:updated_at"`
	PICSChangeNumber       int        `gorm:"not null;column:change_number"`
	PICSConfig             string     `gorm:"not null;column:config"`
	PICSCommon             string     `gorm:"not null;column:common"`
	PICSDepots             string     `gorm:"not null;column:depots"`
	PICSExtended           string     `gorm:"not null;column:extended"`
	PICSRaw                string     `gorm:"not null;column:raw_pics"` // JSON (TEXT)
	ScannedAt              *time.Time `gorm:"not null;column:scanned_at"`
	AchievementPercentages string     `gorm:"not null;column:achievement_percentages;type:text"`
	Achievements           string     `gorm:"not null;column:achievements;type:text"`
	Background             string     `gorm:"not null;column:background"`
	Categories             string     `gorm:"not null;column:categories;type:json"`
	ChangeNumberDate       time.Time  `gorm:"not null;column:change_number_date"`
	ClientIcon             string     `gorm:"not null;column:client_icon"`
	ComingSoon             bool       `gorm:"not null;column:coming_soon"`
	Developers             string     `gorm:"not null;column:developers;type:json"`
	DLC                    string     `gorm:"not null;column:dlc;type:json"`
	DLCCount               int        `gorm:"not null;column:dlc_count"`
	GameID                 int        `gorm:"not null;column:game_id"`
	GameName               string     `gorm:"not null;column:game_name"`
	Genres                 string     `gorm:"not null;column:genres;type:json"`
	HeaderImage            string     `gorm:"not null;column:image_header"`
	Homepage               string     `gorm:"not null;column:homepage"`
	Icon                   string     `gorm:"not null;column:icon"`
	IsFree                 bool       `gorm:"not null;column:is_free;type:tinyint(1)"`
	Logo                   string     `gorm:"not null;column:logo"`
	MetacriticScore        int8       `gorm:"not null;column:metacritic_score"`
	MetacriticURL          string     `gorm:"not null;column:metacritic_url"`
	Movies                 string     `gorm:"not null;column:movies;type:text"`
	Name                   string     `gorm:"not null;column:name"`
	Packages               string     `gorm:"not null;column:packages;type:json"`
	Platforms              string     `gorm:"not null;column:platforms;type:json"`
	Publishers             string     `gorm:"not null;column:publishers;type:json"`
	ReleaseDate            string     `gorm:"not null;column:release_date"`
	ReleaseDateUnix        int64      `gorm:"not null;column:release_date_unix"`
	ReleaseState           string     `gorm:"not null;column:release_state"`
	Schema                 string     `gorm:"not null;column:schema;type:text"`
	Screenshots            string     `gorm:"not null;column:screenshots;type:text"`
	ShortDescription       string     `gorm:"not null;column:description_short"`
	StoreTags              string     `gorm:"not null;column:tags;type:json"`
	Type                   string     `gorm:"not null;column:type"`
	Reviews                string     `gorm:"not null;column:reviews"`
	ReviewsScore           float64    `gorm:"not null;column:reviews_score"`
	ReviewsPositive        int        `gorm:"not null;column:reviews_positive"`
	ReviewsNegative        int        `gorm:"not null;column:reviews_negative"`
	Prices                 string     `gorm:"not null;column:prices"`
	NewsIDs                string     `gorm:"not null;column:news_ids"`
}

func (app *App) BeforeCreate(scope *gorm.Scope) error {

	if app.AchievementPercentages == "" {
		app.AchievementPercentages = "[]"
	}
	if app.Achievements == "" {
		app.Achievements = "{}"
	}
	if app.Categories == "" {
		app.Categories = "[]"
	}
	if app.Developers == "" {
		app.Developers = "[]"
	}
	if app.DLC == "" {
		app.DLC = "[]"
	}
	if app.PICSExtended == "" {
		app.PICSExtended = "{}"
	}
	if app.Prices == "" {
		app.Prices = "{}"
	}
	if app.Genres == "" {
		app.Genres = "[]"
	}
	if app.Movies == "" {
		app.Movies = "[]"
	}
	if app.Packages == "" {
		app.Packages = "[]"
	}
	if app.Platforms == "" {
		app.Platforms = "[]"
	}
	if app.Publishers == "" {
		app.Publishers = "[]"
	}
	if app.Schema == "" {
		app.Schema = "{}"
	}
	if app.Screenshots == "" {
		app.Screenshots = "[]"
	}
	if app.StoreTags == "" {
		app.StoreTags = "[]"
	}

	return nil
}

func (app App) GetID() int {
	return app.ID
}

func (app App) GetProductType() ProductType {
	return ProductTypeApp
}

func (app App) GetPath() string {
	return getAppPath(app.ID, app.Name)
}

func (app App) GetType() (ret string) {

	switch app.Type {
	case "dlc":
		return "DLC"
	case "":
		return "Unknown"
	default:
		return strings.Title(app.Type)
	}
}

func (app App) OutputForJSON(code steam.CountryCode) (output []interface{}) {

	locale, err := helpers.GetLocaleFromCountry(code)
	logging.Error(err)

	return []interface{}{
		app.ID,
		app.GetName(),
		app.GetIcon(),
		app.GetPath(),
		app.GetType(),
		app.ReviewsScore,
		locale.Format(app.GetPrice(code).Final),
		app.UpdatedAt.Unix(),
	}
}

func (app App) OutputForJSONComingSoon(code steam.CountryCode) (output []interface{}) {

	return []interface{}{
		app.ID,
		app.GetName(),
		app.GetIcon(),
		app.GetPath(),
		app.GetType(),
		app.GetPrice(code).GetFinal(code),
		app.GetReleaseDateNice(),
		app.GetReleaseDateUnix(),
	}
}

func (app App) GetDefaultIcon() string {
	return DefaultAppIcon
}

func (app App) GetReleaseState() (ret string) {

	switch app.ReleaseState {
	case "preloadonly":
		return "Preload Only"
	case "prerelease":
		return "Pre Release"
	case "released":
		return "Released"
	case "":
		return "Unreleased"
	default:
		return strings.Title(app.ReleaseState)
	}
}

func (app App) GetReleaseDateNice() string {

	return helpers.GetReleaseDateNice(app.ReleaseDate)
}

func (app App) GetReleaseDateUnix() int64 {

	if app.ReleaseDateUnix == 0 {
		return helpers.GetReleaseDateUnix(app.ReleaseDate)
	}
	return app.ReleaseDateUnix
}

func (app App) GetIcon() (ret string) {

	if app.Icon == "" {
		return app.GetDefaultIcon()
	}
	return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(app.ID) + "/" + app.Icon + ".jpg"
}

func (app *App) SetPrices(prices ProductPrices) (err error) {

	bytes, err := json.Marshal(prices)
	if err != nil {
		return err
	}

	app.Prices = string(bytes)

	return nil
}

func (app App) GetPrices() (prices ProductPrices, err error) {

	err = helpers.Unmarshal([]byte(app.Prices), &prices)
	return prices, err
}

// No erros version, for errors, use GetPrices()
func (app App) GetPrice(code steam.CountryCode) (price ProductPriceCache) {

	prices, err := app.GetPrices()
	if err != nil {
		return price
	}

	price, _ = prices.Get(code)
	return price
}

func (app App) GetNewsIDs() (ids []int64, err error) {

	if app.NewsIDs == "" {
		return ids, err
	}

	err = helpers.Unmarshal([]byte(app.NewsIDs), &ids)
	return ids, err
}

// Adds to current news IDs
func (app *App) SetNewsIDs(news steam.News) (err error) {

	ids, err := app.GetNewsIDs()
	if err != nil {
		return err
	}

	for _, v := range news.Items {
		ids = append(ids, v.GID)
	}

	bytes, err := json.Marshal(helpers.Unique64(ids))
	if err != nil {
		return err
	}

	app.NewsIDs = string(bytes)
	return nil
}

func (app *App) SetReviewScore() {

	if app.ReviewsPositive == 0 && app.ReviewsNegative == 0 {

		app.ReviewsScore = 0

	} else {

		total := float64(app.ReviewsPositive + app.ReviewsNegative)
		average := float64(app.ReviewsPositive) / total
		score := average - (average-0.5)*math.Pow(2, -math.Log10(total + 1))

		app.ReviewsScore = helpers.RoundFloatTo2DP(score * 100)
	}
}

func (app *App) SetExtended(extended PICSExtended) (err error) {

	bytes, err := json.Marshal(extended)
	if err != nil {
		return err
	}

	app.PICSExtended = string(bytes)

	return nil
}

func (app App) GetExtended() (extended PICSExtended, err error) {

	extended = PICSExtended{}

	err = helpers.Unmarshal([]byte(app.PICSExtended), &extended)
	return extended, err
}

func (app *App) SetCommon(common PICSAppCommon) (err error) {

	bytes, err := json.Marshal(common)
	if err != nil {
		return err
	}

	app.PICSCommon = string(bytes)

	return nil
}

func (app App) GetCommon() (common PICSAppCommon, err error) {

	common = PICSAppCommon{}

	err = helpers.Unmarshal([]byte(app.PICSCommon), &common)
	return common, err
}

func (app *App) SetConfig(config PICSAppConfig) (err error) {

	bytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	app.PICSConfig = string(bytes)

	return nil
}

func (app App) GetConfig() (config PICSAppConfig, err error) {

	config = PICSAppConfig{}

	err = helpers.Unmarshal([]byte(app.PICSConfig), &config)
	return config, err
}

func (app *App) SetDepots(depots PICSAppDepots) (err error) {

	bytes, err := json.Marshal(depots)
	if err != nil {
		return err
	}

	app.PICSDepots = string(bytes)

	return nil
}

func (app App) GetDepots() (depots PICSAppDepots, err error) {

	depots = PICSAppDepots{}

	err = helpers.Unmarshal([]byte(app.PICSDepots), &depots)
	return depots, err
}

func (app App) GetCommunityLink() string {
	name := viper.GetString("SHORT_NAME")
	return "https://steamcommunity.com/app/" + strconv.Itoa(app.ID) + "/?utm_source=" + name + "&utm_medium=link&utm_campaign=" + name
}

func (app App) GetStoreLink() string {
	name := viper.GetString("SHORT_NAME")
	return "https://store.steampowered.com/app/" + strconv.Itoa(app.ID) + "/?utm_source=" + name + "&utm_medium=link&utm_campaign=" + name
}

func (app App) GetPCGamingWikiLink() string {
	return "https://pcgamingwiki.com/api/appid.php?appid=" + strconv.Itoa(app.ID)
}

func (app App) GetHeaderImage() string {
	return "http://cdn.akamai.steamstatic.com/steam/apps/" + strconv.Itoa(app.ID) + "/header.jpg"
}

func (app App) GetInstallLink() template.URL {
	return template.URL("steam://install/" + strconv.Itoa(app.ID))
}

func (app App) GetMetacriticLink() template.URL {
	return template.URL("http://www.metacritic.com/game/" + app.MetacriticURL)
}

// Used in template
func (app App) GetScreenshots() (screenshots []steam.AppDetailsScreenshot, err error) {

	err = helpers.Unmarshal([]byte(app.Screenshots), &screenshots)
	return screenshots, err
}

func (app App) GetCoopTags() (string, error) {

	tags, err := app.GetTagIDs()
	if err != nil {
		return "", err
	}

	var tagMap = map[int]string{
		1685: "Co-op",
		3843: "Online co-op",
		3841: "Local co-op",
		4508: "Co-op campaign",

		3859:  "Multiplayer",
		128:   "Massively multiplayer",
		7368:  "Local multiplayer",
		17770: "Asynchronous multiplayer",
	}

	var coopTags []string
	for _, v := range tags {
		if val, ok := tagMap[v]; ok {
			coopTags = append(coopTags, val)
		}
	}

	return strings.Join(coopTags, ", "), nil
}

func (app App) GetAchievements() (achievements steam.AppDetailsAchievements, err error) {

	err = helpers.Unmarshal([]byte(app.Achievements), &achievements)
	return achievements, err
}

func (app App) GetPlatforms() (platforms []string, err error) {

	err = helpers.Unmarshal([]byte(app.Platforms), &platforms)
	return platforms, err
}

func (app App) GetPlatformImages() (ret template.HTML, err error) {

	if app.Platforms == "" {
		return template.HTML(""), nil
	}

	platforms, err := app.GetPlatforms()
	if err != nil {
		return ret, err
	}

	if helpers.SliceHasString(platforms, platformWindows) {
		ret = ret + `<i class="fab fa-windows" data-toggle="tooltip" data-placement="top" title="Windows"></i>`
	} else {
		ret = ret + `<span class="space"></span>`
	}

	if helpers.SliceHasString(platforms, platformMac) {
		ret = ret + `<i class="fab fa-apple" data-toggle="tooltip" data-placement="top" title="Mac"></i>`
	} else {
		ret = ret + `<span class="space"></span>`
	}

	if helpers.SliceHasString(platforms, platformLinux) {
		ret = ret + `<i class="fab fa-linux" data-toggle="tooltip" data-placement="top" title="Linux"></i>`
	} else {
		ret = ret + `<span class="space"></span>`
	}

	return ret, nil
}

func (app App) GetDLC() (dlcs []int, err error) {

	err = helpers.Unmarshal([]byte(app.DLC), &dlcs)
	return dlcs, err
}

func (app App) GetPackages() (packages []int, err error) {

	err = helpers.Unmarshal([]byte(app.Packages), &packages)
	return packages, err
}

func (app App) GetReviews() (reviews steam.ReviewsResponse, err error) {

	err = helpers.Unmarshal([]byte(app.Reviews), &reviews)
	return reviews, err
}

func (app App) GetGenres() (genres []steam.AppDetailsGenre, err error) {

	err = helpers.Unmarshal([]byte(app.Genres), &genres)
	return genres, err
}

func (app App) GetCategories() (categories []string, err error) {

	err = helpers.Unmarshal([]byte(app.Categories), &categories)
	return categories, err
}

func (app App) GetTagIDs() (tags []int, err error) {

	err = helpers.Unmarshal([]byte(app.StoreTags), &tags)
	return tags, err
}

func (app App) GetTags() (tags []Tag, err error) {

	ids, err := app.GetTagIDs()
	if err != nil {
		return tags, err
	}

	return GetTagsByID(ids)
}

func (app App) GetDevelopers() (developers []string, err error) {

	err = helpers.Unmarshal([]byte(app.Developers), &developers)
	return developers, err
}

func (app App) GetPublishers() (publishers []string, err error) {

	err = helpers.Unmarshal([]byte(app.Publishers), &publishers)
	return publishers, err
}

func (app App) GetName() (name string) {
	return getAppName(app.ID, app.Name)
}

// Things that need to happen closer to dailer than when there is an app change
func (app *App) UpdateFromRequest(userAgent string) (errs []error) {

	if !IsValidAppID(app.ID) {
		return []error{ErrInvalidAppID}
	}

	if helpers.IsBot(userAgent) {
		return errs
	}

	if (app.ScannedAt != nil) && (app.ScannedAt.Unix() > (time.Now().Unix() - int64(60*60*24))) { // 1 Days
		return errs
	}

	var err error
	var wg sync.WaitGroup

	// Update news
	wg.Add(1)
	go func(p *App) {

		resp, _, err := helpers.GetSteam().GetNews(app.ID, 1000)
		if err != nil {

			logging.Error(err)

		} else {

			var kinds []Kind
			for _, v := range resp.Items {

				ids, err := app.GetNewsIDs()
				if err != nil {
					errs = append(errs, err)
					continue
				}

				if helpers.SliceHasInt64(ids, v.GID) {
					continue
				}

				kinds = append(kinds, CreateArticle(*app, v))
			}

			err = BulkSaveKinds(kinds, KindNews, false)
			if err != nil {
				errs = append(errs, err)
			}

			err := app.SetNewsIDs(resp)
			logging.Error(err)
		}

		wg.Done()
	}(app)

	// Update reviews
	wg.Add(1)
	go func(p *App) {

		var reviewsResp steam.ReviewsResponse

		reviewsResp, _, err = helpers.GetSteam().GetReviews(app.ID)
		if err != nil {

			errs = append(errs, err)

		} else {

			reviewsBytes, err := json.Marshal(reviewsResp)
			if err != nil {
				errs = append(errs, err)
			}

			app.Reviews = string(reviewsBytes)
			app.ReviewsPositive = reviewsResp.QuerySummary.TotalPositive
			app.ReviewsNegative = reviewsResp.QuerySummary.TotalNegative
			app.SetReviewScore()

			// Log this app score
			err = SaveAppOverTime(*app)
			if err != nil {
				errs = append(errs, err)
			}
		}

		wg.Done()
	}(app)

	// Wait
	wg.Wait()

	// Fix dates
	t := time.Now()
	app.ScannedAt = &t

	// Save
	db, err := GetMySQLClient()
	if err != nil {
		errs = append(errs, err)
	}

	db.Save(app)
	if db.Error != nil {
		errs = append(errs, db.Error)
	}

	return errs
}

// Things that only need to update when there is an app change
func (app *App) UpdateFromAPI() (errs []error) {

	if !IsValidAppID(app.ID) {
		return []error{ErrInvalidAppID}
	}

	var wg sync.WaitGroup

	// Get details from store API
	wg.Add(1)
	go func(app *App) {

		response, _, err := helpers.GetSteam().GetAppDetails(app.ID, steam.CountryUS, steam.LanguageEnglish)
		if err != nil {

			if err == steam.ErrNullResponse {
				errs = append(errs, err)
			}
		}

		// Screenshots
		screenshotsString, err := json.Marshal(response.Data.Screenshots)
		if err != nil {
			errs = append(errs, err)
		}

		// Movies
		moviesString, err := json.Marshal(response.Data.Movies)
		if err != nil {
			errs = append(errs, err)
		}

		// Achievements
		achievementsString, err := json.Marshal(response.Data.Achievements)
		if err != nil {
			errs = append(errs, err)
		}

		// DLC
		dlcString, err := json.Marshal(response.Data.DLC)
		if err != nil {
			errs = append(errs, err)
		}

		// Packages
		packagesString, err := json.Marshal(response.Data.Packages)
		if err != nil {
			errs = append(errs, err)
		}

		// Publishers
		publishersString, err := json.Marshal(response.Data.Publishers)
		if err != nil {
			errs = append(errs, err)
		}

		// Developers
		developersString, err := json.Marshal(response.Data.Developers)
		if err != nil {
			errs = append(errs, err)
		}

		// Categories
		var categories []int8
		for _, v := range response.Data.Categories {
			categories = append(categories, v.ID)
		}

		categoriesString, err := json.Marshal(categories)
		if err != nil {
			errs = append(errs, err)
		}

		genresString, err := json.Marshal(response.Data.Genres)
		if err != nil {
			errs = append(errs, err)
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
			errs = append(errs, err)
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
		app.ReleaseDateUnix = app.GetReleaseDateUnix() // Must be after setting app.ReleaseDate
		app.ComingSoon = response.Data.ReleaseDate.ComingSoon

		// todo, loop through all languages
		prices := ProductPrices{}
		prices.AddPriceFromApp(steam.CountryUS, response)

		app.Prices = prices.String()

		wg.Done()
	}(app)

	// Achievement percentages
	wg.Add(1)
	go func(app *App) {

		percentages, _, err := helpers.GetSteam().GetGlobalAchievementPercentagesForApp(app.ID)
		if err != nil {

			errs = append(errs, err)

		} else {

			percentagesString, err := json.Marshal(percentages)
			if err != nil {

				errs = append(errs, err)

			} else {

				app.AchievementPercentages = string(percentagesString)

			}
		}

		wg.Done()
	}(app)

	// Schema
	wg.Add(1)
	go func(app *App) {

		schema, _, err := helpers.GetSteam().GetSchemaForGame(app.ID)
		if err != nil {

			errs = append(errs, err)

		} else {

			schemaString, err := json.Marshal(schema)
			if err != nil {

				errs = append(errs, err)

			} else {

				app.Schema = string(schemaString)

			}
		}

		wg.Done()
	}(app)

	// Wait
	wg.Wait()

	// Tidy
	app.Type = strings.ToLower(app.Type)
	app.ReleaseState = strings.ToLower(app.ReleaseState)

	return errs
}

func GetTypesForSelect() []AppType {

	types := []string{
		"game",
		"advertising",
		"application",
		"config",
		"demo",
		"dlc",
		"episode",
		"guide",
		"hardware",
		"media",
		"mod",
		"movie",
		"series",
		"tool",
		"", // Displays as Unknown
		"video",
	}

	var ret []AppType
	for _, v := range types {
		ret = append(ret, AppType{
			ID:   v,
			Name: App{Type: v}.GetType(),
		})
	}

	return ret
}

type AppType struct {
	ID   string
	Name string
}

func GetApp(id int) (app App, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return app, err
	}

	db.First(&app, id)
	if db.Error != nil {
		return app, db.Error
	}

	if app.ID == 0 {
		return app, ErrCantFindApp
	}

	return app, nil
}

func GetAppsByID(ids []int, columns []string) (apps []App, err error) { // todo, chunk ids into multple queries async

	if len(ids) == 0 {
		return apps, nil
	}

	db, err := GetMySQLClient()
	if err != nil {
		return apps, err
	}

	if len(columns) > 0 {
		db = db.Select(columns)
	}

	db.Where("id IN (?)", ids).Find(&apps)
	if db.Error != nil {
		return apps, db.Error
	}

	return apps, nil
}

// todo, these methods could all be one?
func GetAppsWithTags() (apps []App, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return apps, err
	}

	db = db.Select([]string{"tags", "prices", "reviews_score"})
	db = db.Where("JSON_DEPTH(tags) = 2")
	db = db.Order("id asc")

	db = db.Find(&apps)
	if db.Error != nil {
		return apps, db.Error
	}

	return apps, nil
}

func GetAppsWithDevelopers() (apps []App, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return apps, err
	}

	db = db.Select([]string{"developers", "prices", "reviews_score"})
	db = db.Where("JSON_DEPTH(developers) = 2")

	db = db.Find(&apps)
	if db.Error != nil {
		return apps, db.Error
	}

	return apps, nil
}

func GetAppsWithPublishers() (apps []App, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return apps, err
	}

	db = db.Select([]string{"publishers", "prices", "reviews_score"})
	db = db.Where("JSON_DEPTH(publishers) = 2")

	db = db.Find(&apps)
	if db.Error != nil {
		return apps, db.Error
	}

	return apps, nil
}

func GetAppsWithGenres() (apps []App, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return apps, err
	}

	db = db.Select([]string{"genres", "prices", "reviews_score"})
	db = db.Where("JSON_DEPTH(genres) = 3")

	db = db.Find(&apps)
	if db.Error != nil {
		return apps, db.Error
	}

	return apps, nil
}

func GetDLC(app App, columns []string) (apps []App, err error) {

	dlc, err := app.GetDLC()
	if err != nil {
		return apps, err
	}

	if len(dlc) == 0 {
		return apps, nil
	}

	db, err := GetMySQLClient()
	if err != nil {
		return apps, err
	}

	db = db.Where("id in (?)", dlc).Find(&apps)

	if len(columns) > 0 {
		db = db.Select(columns)
	}

	return apps, db.Error
}

func CountApps() (count int, err error) {

	return helpers.GetMemcache().GetSetInt(helpers.MemcacheAppsCount, func() (count int, err error) {

		db, err := GetMySQLClient()
		if err != nil {
			return count, err
		}

		db.Model(&App{}).Count(&count)
		return count, db.Error
	})
}

func GetMostExpensiveApp(code steam.CountryCode) (price int, err error) {

	return helpers.GetMemcache().GetSetInt(helpers.MemcacheMostExpensiveApp(code), func() (count int, err error) {

		db, err := GetMySQLClient()
		if err != nil {
			return count, err
		}

		var countSlice []int
		db.Model(&App{}).Pluck("max(prices->\"$."+string(code)+".final\")", &countSlice)
		if db.Error != nil {
			return count, db.Error
		}
		if len(countSlice) != 1 {
			return count, errors.New("query failed")
		}

		return countSlice[0], nil
	})
}

func IsValidAppID(id int) bool {
	return id != 0
}

func getAppPath(id int, name string) string {

	p := "/games/" + strconv.Itoa(id)

	if name != "" {
		p = p + "/" + slug.Make(name)
	}

	return p
}

func getAppName(id int, name string) string {

	if name != "" {
		return name
	} else if id > 0 {
		return "App " + strconv.Itoa(id)
	}
	return "Unknown App"
}
