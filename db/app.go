package db

import (
	"encoding/json"
	"errors"
	"html/template"
	"math"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gosimple/slug"
	"github.com/spf13/viper"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/memcache"
	"github.com/steam-authority/steam-authority/steami"
)

const (
	platformWindows = "windows"
	platformMac     = "macos"
	platformLinux   = "linux"
)

var (
	ErrInvalidAppID = UpdateError{"invalid app id", true, false}
	ErrCantFindApp  = errors.New("no such app")
)

type App struct {
	ID                     int        `gorm:"not null;column:id;primary_key"`
	CreatedAt              *time.Time `gorm:"not null;column:created_at"`
	UpdatedAt              *time.Time `gorm:"not null;column:updated_at"`
	ScannedAt              *time.Time `gorm:"not null;column:scanned_at"`
	AchievementPercentages string     `gorm:"not null;column:achievement_percentages;type:text;default:'[]'"`
	Achievements           string     `gorm:"not null;column:achievements;type:text;default:'{}'"`
	Background             string     `gorm:"not null;column:background"`
	Categories             string     `gorm:"not null;column:categories;type:json;default:'[]'"`
	ChangeNumber           int        `gorm:"not null;column:change_number"`
	ClientIcon             string     `gorm:"not null;column:client_icon"`
	ComingSoon             bool       `gorm:"not null;column:coming_soon"`
	Developers             string     `gorm:"not null;column:developers;type:json;default:'[]'"`
	DLC                    string     `gorm:"not null;column:dlc;type:json;default:'[]'"`
	DLCCount               int        `gorm:"not null;column:dlc_count"`
	Extended               string     `gorm:"not null;column:extended;default:'{}'"`
	GameID                 int        `gorm:"not null;column:game_id"`
	GameName               string     `gorm:"not null;column:game_name"`
	Genres                 string     `gorm:"not null;column:genres;type:json;default:'[]'"`
	HeaderImage            string     `gorm:"not null;column:image_header"`
	Homepage               string     `gorm:"not null;column:homepage"`
	Icon                   string     `gorm:"not null;column:icon"`
	IsFree                 bool       `gorm:"not null;column:is_free;type:tinyint(1)"`
	Logo                   string     `gorm:"not null;column:logo"`
	MetacriticScore        int8       `gorm:"not null;column:metacritic_score"`
	MetacriticURL          string     `gorm:"not null;column:metacritic_url"`
	Movies                 string     `gorm:"not null;column:movies;type:text;default:'[]'"`
	Name                   string     `gorm:"not null;column:name"`
	Packages               string     `gorm:"not null;column:packages;type:json;default:'[]'"`
	Platforms              string     `gorm:"not null;column:platforms;type:json;default:'[]'"`
	PriceDiscount          int        `gorm:"not null;column:price_discount"`
	PriceFinal             int        `gorm:"not null;column:price_final"`
	PriceInitial           int        `gorm:"not null;column:price_initial"`
	Publishers             string     `gorm:"not null;column:publishers;type:json;default:'[]'"`
	ReleaseDate            string     `gorm:"not null;column:release_date"`
	ReleaseState           string     `gorm:"not null;column:release_state"`
	Schema                 string     `gorm:"not null;column:schema;type:text;default:'{}'"`
	Screenshots            string     `gorm:"not null;column:screenshots;type:text;default:'[]'"`
	ShortDescription       string     `gorm:"not null;column:description_short"`
	StoreTags              string     `gorm:"not null;column:tags;type:json;default:'[]'"`
	Type                   string     `gorm:"not null;column:type"`
	Reviews                string     `gorm:"not null;column:reviews"`
	ReviewsScore           float64    `gorm:"not null;column:reviews_score"`
	ReviewsPositive        int        `gorm:"not null;column:reviews_positive"`
	ReviewsNegative        int        `gorm:"not null;column:reviews_negative"`
}

//func GetDefaultAppJSON() App {
//	return App{
//		AchievementPercentages: "[]",
//		Achievements:           "{}",
//		Categories:             "[]",
//		Developers:             "[]",
//		DLC:                    "[]",
//		Extended:               "{}",
//		Genres:                 "[]",
//		Movies:                 "[]",
//		Packages:               "[]",
//		Platforms:              "[]",
//		Publishers:             "[]",
//		Schema:                 "{}",
//		Screenshots:            "[]",
//		StoreTags:              "[]",
//	}
//}

func (app App) GetPath() string {
	return getAppPath(app.ID, app.GetName())
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

func GetTypesForSelect() (ret map[string]string) {

	types := []string{
		"",
		"advertising",
		"application",
		"config",
		"demo",
		"dlc",
		"episode",
		"game",
		"guide",
		"hardware",
		"media",
		"mod",
		"movie",
		"series",
		"tool",
		"video",
	}

	m := map[string]string{}
	for _, v := range types {
		m[v] = App{Type: v}.GetType()
	}
	return m
}

func (app App) OutputForJSON() (output []interface{}) {

	return []interface{}{
		app.ID,
		app.GetName(),
		app.GetIcon(),
		app.GetPath(),
		app.GetType(),
		app.ReviewsScore,
		app.DLCCount,
	}
}

func (app App) GetDefaultAvatar() string {
	return "/assets/img/no-app-image-square.jpg"
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

	return helpers.GetReleaseDateUnix(app.ReleaseDate)
}

func (app App) GetIcon() (ret string) {

	if app.Icon == "" {
		return "/assets/img/no-app-image-square.jpg"
	}
	return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(app.ID) + "/" + app.Icon + ".jpg"
}

func (app App) GetPriceInitial() float64 {
	return helpers.CentsInt(app.PriceInitial)
}

func (app App) GetPriceFinal() float64 {
	return helpers.CentsInt(app.PriceFinal)
}

func (app App) GetPriceFinalNice() string {

	if app.PriceFinal == 0 {
		return "Free"
	}
	return "$" + strconv.FormatFloat(app.GetPriceFinal(), 'f', 2, 64)
}

func (app App) GetReviewScore() float64 {

	if app.ReviewsPositive == 0 && app.ReviewsNegative == 0 {
		return 0
	}

	total := float64(app.ReviewsPositive + app.ReviewsNegative)
	average := float64(app.ReviewsPositive) / total
	score := average - (average-0.5)*math.Pow(2, -math.Log10(total + 1))

	return helpers.DollarsFloat(score * 100)
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

func (app App) GetCoopTags(tagMap map[int]string) string {

	tags, err := app.GetTagIDs()
	if err != nil {
		logger.Error(err)
		return ""
	}

	var coopTags []string
	for _, v := range tags {
		if val, ok := tagMap[v]; ok {
			coopTags = append(coopTags, val)
		}
	}

	return strings.Join(coopTags, ", ")
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

		var articles []*News

		articles, err = GetNewArticles(app.ID)
		if err != nil {

			errs = append(errs, err)

		} else {

			var kinds []Kind
			for _, v := range articles {
				kinds = append(kinds, *v)
			}

			err = BulkSaveKinds(kinds, KindNews, true)
			if err != nil {
				errs = append(errs, err)
			}
		}

		wg.Done()
	}(app)

	// Update reviews
	wg.Add(1)
	go func(p *App) {

		var reviewsResp steam.ReviewsResponse

		reviewsResp, _, err = steami.Steam().GetReviews(app.ID)
		if err != nil {

			errs = append(errs, err)

		} else {

			reviewsBytes, err := json.Marshal(reviewsResp)
			if err != nil {
				errs = append(errs, err)
			}

			app.Reviews = string(reviewsBytes)
			app.ReviewsScore = app.GetReviewScore()
			app.ReviewsPositive = reviewsResp.QuerySummary.TotalPositive
			app.ReviewsNegative = reviewsResp.QuerySummary.TotalNegative

			// Log this app score
			err = SaveAppReviewScore(app.ID, app.ReviewsScore, app.ReviewsPositive, app.ReviewsNegative)
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
		errs = append(errs, err)
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

		response, _, err := steami.Steam().GetAppDetails(app.ID)
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
		app.ComingSoon = response.Data.ReleaseDate.ComingSoon
		app.PriceInitial = response.Data.PriceOverview.Initial
		app.PriceFinal = response.Data.PriceOverview.Final
		app.PriceDiscount = response.Data.PriceOverview.DiscountPercent

		wg.Done()
	}(app)

	// Achievement percentages
	wg.Add(1)
	go func(app *App) {

		percentages, _, err := steami.Steam().GetGlobalAchievementPercentagesForApp(app.ID)
		if err != nil {

			logger.Error(err)

		} else {

			percentagesString, err := json.Marshal(percentages)
			if err != nil {

				logger.Error(err)

			} else {

				app.AchievementPercentages = string(percentagesString)

			}
		}

		wg.Done()
	}(app)

	// Schema
	wg.Add(1)
	go func(app *App) {

		schema, _, err := steami.Steam().GetSchemaForGame(app.ID)
		if err != nil {

			logger.Error(err)

		} else {

			schemaString, err := json.Marshal(schema)
			if err != nil {

				logger.Error(err)

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

	// Default JSON values
	if app.StoreTags == "" || app.StoreTags == "null" {
		app.StoreTags = "[]"
	}

	if app.Categories == "" || app.Categories == "null" {
		app.Categories = "[]"
	}

	if app.Genres == "" || app.Genres == "null" {
		app.Genres = "[]"
	}

	if app.Screenshots == "" || app.Screenshots == "null" {
		app.Screenshots = "[]"
	}

	if app.Movies == "" || app.Movies == "null" {
		app.Movies = "[]"
	}

	if app.Achievements == "" || app.Achievements == "null" {
		app.Achievements = "{}"
	}

	if app.Platforms == "" || app.Platforms == "null" {
		app.Platforms = "[]"
	}

	if app.DLC == "" || app.DLC == "null" {
		app.DLC = "[]"
	}

	if app.Packages == "" || app.Packages == "null" {
		app.Packages = "[]"
	}

	if app.AchievementPercentages == "" || app.AchievementPercentages == "null" {
		app.AchievementPercentages = "[]"
	}

	if app.Developers == "" || app.Developers == "null" {
		app.Developers = "[]"
	}

	if app.Publishers == "" || app.Publishers == "null" {
		app.Publishers = "[]"
	}

	if app.Schema == "" || app.Schema == "null" {
		app.Schema = "{}"
	}

	if app.Extended == "" || app.Extended == "null" {
		app.Extended = "{}"
	}

	return errs
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

	if len(ids) < 1 {
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

func SearchApps(query url.Values, limit int, page int, sort string, columns []string) (apps []App, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return apps, err
	}

	if limit > 0 {
		db = db.Limit(limit)
	}

	offset := (page - 1) * limit
	db = db.Offset(offset)

	if sort != "" {
		db = db.Order(sort)
	}

	if len(columns) > 0 {
		db = db.Select(columns)
	}

	// Type
	if _, ok := query["type"]; ok {
		db = db.Where("type = ?", query.Get("type"))
	}

	// Tags depth
	if _, ok := query["tags_depth"]; ok {
		db = db.Where("JSON_DEPTH(tags) = ?", query.Get("tags_depth"))
	}

	// Genres depth
	if _, ok := query["genres_depth"]; ok {
		db = db.Where("JSON_DEPTH(genres) = ?", query.Get("genres_depth"))
	}

	// Free
	if _, ok := query["is_free"]; ok {
		db = db.Where("is_free = ?", query.Get("is_free"))
	}

	// Platforms
	if _, ok := query["platforms"]; ok {
		db = db.Where("JSON_CONTAINS(platforms, [\"?\"])", query.Get("platforms"))

	}

	// Tag
	if _, ok := query["tags"]; ok {
		db = db.Where("JSON_CONTAINS(tags, ?)", "[\""+query.Get("tags")+"\"]")
	}

	// Genres
	// select * from apps WHERE JSON_SEARCH(genres, 'one', 'Action') IS NOT NULL;

	// Query
	db = db.Find(&apps)
	if db.Error != nil {
		return apps, db.Error
	}

	return apps, err
}

func GetDLC(app App, columns []string) (apps []App, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return apps, err
	}

	dlc, err := app.GetDLC()
	if err != nil {
		return apps, err
	}

	db = db.Where("id in (?)", dlc).Find(&apps)

	if len(columns) > 0 {
		db = db.Select(columns)
	}

	if db.Error != nil {
		return apps, db.Error
	}

	return apps, nil
}

func CountApps() (count int, err error) {

	return memcache.GetSetInt(memcache.AppsCount, func() (count int, err error) {

		db, err := GetMySQLClient()
		if err != nil {
			return count, err
		}

		db.Model(&App{}).Count(&count)
		return count, db.Error
	})
}

func IsValidAppID(id int) bool {

	if id == 0 {
		return false
	}

	return true
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
