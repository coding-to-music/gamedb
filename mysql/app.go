package mysql

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/go-helpers/logger"
	"github.com/gosimple/slug"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/steam-authority/steam-authority/steam"
)

type App struct {
	ID                     int        `gorm:"not null;column:id;primary_key;AUTO_INCREMENT"`                  //
	CreatedAt              *time.Time `gorm:"not null;column:created_at"`                                     //
	UpdatedAt              *time.Time `gorm:"not null;column:updated_at"`                                     //
	Name                   string     `gorm:"not null;column:name"`                                           //
	Type                   string     `gorm:"not null;column:type"`                                           //
	IsFree                 bool       `gorm:"not null;column:is_free;type:tinyint(1)"`                        //
	DLC                    string     `gorm:"not null;column:dlc;type:json;default:'[]'"`                     // JSON
	ShortDescription       string     `gorm:"not null;column:description_short"`                              //
	HeaderImage            string     `gorm:"not null;column:image_header"`                                   //
	Developer              string     `gorm:"not null;column:developer"`                                      //
	Publisher              string     `gorm:"not null;column:publisher"`                                      //
	Packages               string     `gorm:"not null;column:packages;type:json;default:'[]'"`                // JSON
	MetacriticScore        int8       `gorm:"not null;column:metacritic_score"`                               //
	MetacriticURL          string     `gorm:"not null;column:metacritic_url"`                                 //
	Categories             string     `gorm:"not null;column:categories;type:json;default:'[]'"`              // JSON
	Genres                 string     `gorm:"not null;column:genres;type:json;default:'[]'"`                  // JSON
	Screenshots            string     `gorm:"not null;column:screenshots;type:text;default:'[]'"`             // JSON
	Movies                 string     `gorm:"not null;column:movies;type:text;default:'[]'"`                  // JSON
	Achievements           string     `gorm:"not null;column:achievements;type:text;default:'{}'"`            // JSON
	Background             string     `gorm:"not null;column:background"`                                     //
	Platforms              string     `gorm:"not null;column:platforms;type:json;default:'[]'"`               // JSON
	GameID                 int        `gorm:"not null;column:game_id"`                                        //
	GameName               string     `gorm:"not null;column:game_name"`                                      //
	ReleaseState           string     `gorm:"not null;column:release_state"`                                  // PICS
	StoreTags              string     `gorm:"not null;column:tags;type:json;default:'[]'"`                    // PICS JSON
	Homepage               string     `gorm:"not null;column:homepage"`                                       // PICS
	ChangeNumber           int        `gorm:"not null;column:change_number"`                                  // PICS
	Logo                   string     `gorm:"not null;column:logo"`                                           // PICS
	Icon                   string     `gorm:"not null;column:icon"`                                           // PICS
	ClientIcon             string     `gorm:"not null;column:client_icon"`                                    // PICS
	Ghost                  bool       `gorm:"not null;column:is_ghost;type:tinyint(1)"`                       //
	PriceInitial           int        `gorm:"not null;column:price_initial"`                                  //
	PriceFinal             int        `gorm:"not null;column:price_final"`                                    //
	PriceDiscount          int        `gorm:"not null;column:price_discount"`                                 //
	AchievementPercentages string     `gorm:"not null;column:achievement_percentages;type:text;default:'[]'"` // JSON
	Schema                 string     `gorm:"not null;column:schema;type:text;default:'{}'"`                  // JSON
	ComingSoon             bool       `gorm:"not null;column:coming_soon"`                                    //
	ReleaseDate            string     `gorm:"not null;column:release_date"`                                   //
}

func GetDefaultAppJSON() App {
	return App{
		StoreTags:              "[]",
		Categories:             "[]",
		Genres:                 "[]",
		Screenshots:            "[]",
		Movies:                 "[]",
		Achievements:           "{}",
		Platforms:              "[]",
		DLC:                    "[]",
		Packages:               "[]",
		AchievementPercentages: "[]",
		Schema:                 "{}",
	}
}

func (app App) GetPath() (ret string) {
	ret = "/apps/" + strconv.Itoa(int(app.ID))

	if app.Name != "" {
		ret = ret + "/" + slug.Make(app.Name)
	}

	return ret
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

func (app App) GetIcon() (ret string) {

	if app.Icon == "" {
		return "/assets/img/no-app-image-square.jpg"
	} else {
		return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(app.ID) + "/" + app.Icon + ".jpg"
	}
}

func (app App) GetPriceInitial() string {
	return fmt.Sprintf("%0.2f", float64(app.PriceInitial)/100)
}

func (app App) GetCommunityLink() (string) {
	return "https://steamcommunity.com/app/" + strconv.Itoa(app.ID) + "/?utm_source=SteamAuthority&utm_medium=link&utm_campaign=SteamAuthority"
}

func (app App) GetStoreLink() (string) {
	return "https://store.steampowered.com/app/" + strconv.Itoa(app.ID) + "/?utm_source=SteamAuthority&utm_medium=link&utm_campaign=SteamAuthority"
}

func (app App) GetInstallLink() (template.URL) {
	return template.URL("steam://install/" + strconv.Itoa(app.ID))
}

func (app App) GetMetacriticLink() (template.URL) {
	return template.URL("http://www.metacritic.com/game/" + app.MetacriticURL)
}

// Used in frontend
func (app App) GetScreenshots() (screenshots []steam.AppDetailsScreenshot, err error) {

	bytes := []byte(app.Screenshots)
	if err := json.Unmarshal(bytes, &screenshots); err != nil {
		return screenshots, err
	}

	return screenshots, nil
}

// Used in template
func (app App) GetCoopTags(tagMap map[int]string) (string) {

	tags, err := app.GetTags()
	if err != nil {
		logger.Error(err)
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

	bytes := []byte(app.Achievements)
	if err := json.Unmarshal(bytes, &achievements); err != nil {
		return achievements, err
	}

	return achievements, nil
}

func (app App) GetPlatforms() (platforms []string, err error) {

	bytes := []byte(app.Platforms)
	if len(bytes) == 0 {
		return platforms, nil
	}

	if err := json.Unmarshal(bytes, &platforms); err != nil {
		return platforms, err
	}

	return platforms, nil
}

func (app App) GetPlatformImages() (ret template.HTML, err error) {

	platforms, err := app.GetPlatforms()
	if err != nil {
		return ret, err
	}

	for _, v := range platforms {
		if v == "macos" {
			ret = ret + `<i class="fab fa-apple"></i>`
		} else if v == "windows" {
			ret = ret + `<i class="fab fa-windows"></i>`
		} else if v == "linux" {
			ret = ret + `<i class="fab fa-linux"></i>`
		}
	}

	return ret, nil
}

func (app App) GetDLC() (dlcs []int, err error) {

	bytes := []byte(app.DLC)
	if err := json.Unmarshal(bytes, &dlcs); err != nil {
		return dlcs, err
	}

	return dlcs, nil
}

func (app App) GetPackages() (packages []int, err error) {

	bytes := []byte(app.Packages)
	if err := json.Unmarshal(bytes, &packages); err != nil {
		return packages, err
	}

	return packages, nil
}

func (app App) GetGenres() (genres []steam.AppDetailsGenre, err error) {

	bytes := []byte(app.Genres)
	if err := json.Unmarshal(bytes, &genres); err != nil {
		return genres, err
	}

	return genres, nil
}

func (app App) GetCategories() (categories []string, err error) {

	bytes := []byte(app.Categories)
	if err := json.Unmarshal(bytes, &categories); err != nil {
		return categories, err
	}

	return categories, nil
}

func (app App) GetTags() (tags []int, err error) {

	bytes := []byte(app.StoreTags)
	if err := json.Unmarshal(bytes, &tags); err != nil {
		return tags, err
	}

	return tags, nil
}

func (app App) GetName() (name string) {

	if app.Name == "" {
		app.Name = "App " + strconv.Itoa(app.ID)
	}

	return app.Name
}

func GetApp(id int) (app App, err error) {

	db, err := GetDB()
	if err != nil {
		return app, err
	}

	db.First(&app, id)
	if db.Error != nil {
		return app, db.Error
	}

	if app.ID == 0 {
		return app, errors.New("no id")
	}

	return app, nil
}

func GetApps(ids []int, columns []string) (apps []App, err error) {

	if len(ids) < 1 {
		return apps, nil
	}

	db, err := GetDB()
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

func SearchApps(query url.Values, limit int, sort string, columns []string) (apps []App, err error) {

	db, err := GetDB()
	if err != nil {
		return apps, err
	}

	if limit > 0 {
		db = db.Limit(limit)
	}

	if sort != "" {
		db = db.Order(sort)
	}

	if len(columns) > 0 {
		db = db.Select(columns)
	}

	// Hide ghosts? todo, fix, apps can have no name
	db = db.Where("name != ''")

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

func CountApps() (count int, err error) {

	db, err := GetDB()
	if err != nil {
		return count, err
	}

	db.Model(&App{}).Count(&count)
	if db.Error != nil {
		return count, db.Error
	}

	return count, nil
}

func (app *App) Fill() (err error) {

	// Get app details
	err = app.fillFromAPI()
	if err != nil {
		return err
	}

	// PICS
	err = app.fillFromPICS()
	if err != nil {
		if err.Error() != "no app key in json" {
			return err
		}
	}

	// Achievement percentages
	percentages, err := steam.GetGlobalAchievementPercentagesForApp(app.ID)
	if err != nil {
		logger.Error(err)
	}

	percentagesString, err := json.Marshal(percentages)
	if err != nil {
		logger.Error(err)
	}

	app.AchievementPercentages = string(percentagesString)

	// Schema
	schema, err := steam.GetSchemaForGame(app.ID)
	if err != nil {
		logger.Error(err)
	}

	schemaString, err := json.Marshal(schema)
	if err != nil {
		logger.Error(err)
	}

	app.Schema = string(schemaString)

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

	if app.Schema == "" || app.Schema == "null" {
		app.Schema = "{}"
	}

	return nil
}

func (app *App) fillFromAPI() (err error) {

	// Get data
	response, err := steam.GetAppDetailsFromStore(app.ID)
	if err != nil {

		// Not all apps can be found
		if err.Error() == "no app with id in steam" || strings.HasPrefix(err.Error(), "invalid app id:") {
			return nil
		}

		return err
	}

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
	app.ShortDescription = response.Data.ShortDescription
	app.HeaderImage = response.Data.HeaderImage
	app.Developer = strings.Join(response.Data.Developers, ", ")
	app.Publisher = strings.Join(response.Data.Publishers, ", ")
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

	return nil
}

func (app *App) fillFromPICS() (err error) {

	// Call PICS
	resp, err := steam.GetPICSInfo([]int{app.ID}, []int{})
	if err != nil {
		return err
	}

	var js steam.JsApp
	if len(resp.Apps) > 0 {
		js = resp.Apps[strconv.Itoa(app.ID)]
	} else {
		return errors.New("no app key in json")
	}

	// Check if empty
	app.Ghost = reflect.DeepEqual(js.Common, steam.JsAppCommon{})

	// Tags, convert map to slice
	var tagsSlice []int
	for _, v := range js.Common.StoreTags {
		vv, _ := strconv.Atoi(v)
		tagsSlice = append(tagsSlice, vv)
	}

	tags, err := json.Marshal(tagsSlice)
	if err != nil {
		return err
	}

	// Meta critic
	var metacriticScoreInt = 0
	if js.Common.MetacriticScore != "" {
		metacriticScoreInt, _ = strconv.Atoi(js.Common.MetacriticScore)
	}

	//
	app.Name = js.Common.Name
	app.Type = js.Common.Type
	app.ReleaseState = js.Common.ReleaseState
	// app.Platforms = strings.Split(js.Common.OSList, ",") // Can get from API
	app.MetacriticScore = int8(metacriticScoreInt)
	app.MetacriticURL = js.Common.MetacriticURL
	app.StoreTags = string(tags)
	app.Developer = js.Extended.Developer
	app.Publisher = js.Extended.Publisher
	app.Homepage = js.Extended.Homepage
	app.ChangeNumber = js.ChangeNumber
	app.Logo = js.Common.Logo
	app.Icon = js.Common.Icon
	app.ClientIcon = js.Common.ClientIcon

	return nil
}
