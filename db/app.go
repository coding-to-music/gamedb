package db

import (
	"errors"
	"html/template"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
)

const (
	platformWindows = "windows"
	platformMac     = "macos"
	platformLinux   = "linux"

	DefaultAppIcon = "/assets/img/no-app-image-square.jpg"
)

type App struct {
	Achievements       string    `gorm:"not null;column:achievements;type:text"`           // []AppAchievement
	Background         string    `gorm:"not null;column:background"`                       //
	BundleIDs          string    `gorm:"not null;column:bundle_ids"`                       //
	Categories         string    `gorm:"not null;column:categories;type:json"`             //
	ChangeNumber       int       `gorm:"not null;column:change_number"`                    //
	ChangeNumberDate   time.Time `gorm:"not null;column:change_number_date;type:datetime"` //
	ClientIcon         string    `gorm:"not null;column:client_icon"`                      //
	ComingSoon         bool      `gorm:"not null;column:coming_soon"`                      //
	Common             string    `gorm:"not null;column:common"`                           //
	Config             string    `gorm:"not null;column:config"`                           //
	CreatedAt          time.Time `gorm:"not null;column:created_at;type:datetime"`         //
	Depots             string    `gorm:"not null;column:depots"`                           //
	Developers         string    `gorm:"not null;column:developers;type:json"`             //
	DLC                string    `gorm:"not null;column:dlc;type:json"`                    //
	DLCCount           int       `gorm:"not null;column:dlc_count"`                        //
	Extended           string    `gorm:"not null;column:extended"`                         //
	GameID             int       `gorm:"not null;column:game_id"`                          //
	GameName           string    `gorm:"not null;column:game_name"`                        //
	Genres             string    `gorm:"not null;column:genres;type:json"`                 // []AppGenre
	HeaderImage        string    `gorm:"not null;column:image_header"`                     //
	Homepage           string    `gorm:"not null;column:homepage"`                         //
	Icon               string    `gorm:"not null;column:icon"`                             //
	ID                 int       `gorm:"not null;column:id;primary_key"`                   //
	Install            string    `gorm:"not null;column:install"`                          //
	IsFree             bool      `gorm:"not null;column:is_free;type:tinyint(1)"`          //
	Launch             string    `gorm:"not null;column:launch"`                           //
	Localization       string    `gorm:"not null;column:localization"`                     //
	Logo               string    `gorm:"not null;column:logo"`                             //
	MetacriticScore    int8      `gorm:"not null;column:metacritic_score"`                 //
	MetacriticURL      string    `gorm:"not null;column:metacritic_url"`                   //
	Movies             string    `gorm:"not null;column:movies;type:text"`                 // []AppVideo
	Name               string    `gorm:"not null;column:name"`                             //
	NewsIDs            string    `gorm:"not null;column:news_ids"`                         //
	Packages           string    `gorm:"not null;column:packages;type:json"`               // []int
	Platforms          string    `gorm:"not null;column:platforms;type:json"`              //
	Prices             string    `gorm:"not null;column:prices"`                           //
	PublicOnly         bool      `gorm:"not null;column:public_only"`                      //
	Publishers         string    `gorm:"not null;column:publishers;type:json"`             //
	ReleaseDate        string    `gorm:"not null;column:release_date"`                     //
	ReleaseDateUnix    int64     `gorm:"not null;column:release_date_unix"`                //
	ReleaseState       string    `gorm:"not null;column:release_state"`                    //
	Reviews            string    `gorm:"not null;column:reviews"`                          // AppReviewSummary
	ReviewsScore       float64   `gorm:"not null;column:reviews_score"`                    //
	Screenshots        string    `gorm:"not null;column:screenshots;type:text"`            // []AppImage
	ShortDescription   string    `gorm:"not null;column:description_short"`                //
	Stats              string    `gorm:"not null;column:stats;type:text"`                  // []AppStat
	SteamSpy           string    `gorm:"not null;column:steam_spy"`                        // AppSteamSpy
	StoreTags          string    `gorm:"not null;column:tags;type:json"`                   //
	SystemRequirements string    `gorm:"not null;column:system_requirements"`              //
	Type               string    `gorm:"not null;column:type"`                             //
	UFS                string    `gorm:"not null;column:ufs"`                              //
	UpdatedAt          time.Time `gorm:"not null;column:updated_at;type:datetime"`         //
	Version            string    `gorm:"not null;column:version"`                          //
}

func (app *App) BeforeSave(scope *gorm.Scope) error {

	if app.Achievements == "" {
		app.Achievements = "[]"
	}
	if app.BundleIDs == "" {
		app.BundleIDs = "[]"
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
	if app.Extended == "" {
		app.Extended = "{}"
	}
	if app.SystemRequirements == "" {
		app.SystemRequirements = "{}"
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
	if app.Stats == "" {
		app.Stats = "[]"
	}
	if app.Screenshots == "" {
		app.Screenshots = "[]"
	}
	if app.StoreTags == "" {
		app.StoreTags = "[]"
	}
	if app.SteamSpy == "" {
		app.SteamSpy = "{}"
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
	return GetAppPath(app.ID, app.Name)
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

func (app App) GetDaysToRelease() string {

	return helpers.GetDaysToRelease(app.ReleaseDateUnix)
}

func (app App) OutputForJSON(code steam.CountryCode) (output []interface{}) {

	var cnds string
	cnd := app.ChangeNumberDate.Unix()
	if cnd == -62167219200 {
		cnds = "Unknown"
	} else {
		cnds = strconv.FormatInt(app.ChangeNumberDate.Unix(), 10)
	}

	return []interface{}{
		app.ID,
		app.GetName(),
		app.GetIcon(),
		app.GetPath(),
		app.GetType(),
		app.ReviewsScore,
		GetPriceFormatted(app, code).Final,
		cnds,
	}
}

// Must be the same as package OutputForJSONUpcoming
func (app App) OutputForJSONUpcoming(code steam.CountryCode) (output []interface{}) {

	return []interface{}{
		app.ID,
		app.GetName(),
		app.GetIcon(),
		app.GetPath(),
		app.GetType(),
		GetPriceFormatted(app, code).Final,
		app.GetDaysToRelease(),
		app.GetReleaseDateNice(),
	}
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

	if app.ReleaseDateUnix == 0 {
		return app.ReleaseDate
	}

	return time.Unix(app.ReleaseDateUnix, 0).Format(helpers.DateYear)
}

func (app App) GetUpdatedNice() string {
	return app.UpdatedAt.Format(helpers.DateYearTime)
}

func (app App) GetPICSUpdatedNice() string {

	d := app.ChangeNumberDate

	// Empty dates
	if d.IsZero() || d.Unix() == -62167219200 {
		return "-"
	}
	return d.Format(helpers.DateYearTime)
}

func (app App) GetIcon() (ret string) {
	return GetAppIcon(app.ID, app.Icon)
}

func (app App) GetPrices() (prices ProductPrices, err error) {

	err = helpers.Unmarshal([]byte(app.Prices), &prices)
	log.Err(err)
	return prices, err
}

func (app App) GetPrice(code steam.CountryCode) (price ProductPriceStruct, err error) {

	prices, err := app.GetPrices()
	if err != nil {
		return price, err
	}

	return prices.Get(code)
}

func (app App) GetNewsIDs() (ids []int64, err error) {

	if app.NewsIDs == "" {
		return ids, err
	}

	err = helpers.Unmarshal([]byte(app.NewsIDs), &ids)
	return ids, err
}

func (app App) GetExtended() (extended PICSExtended, err error) {

	extended = PICSExtended{}

	err = helpers.Unmarshal([]byte(app.Extended), &extended)
	log.Err(err)
	return extended, err
}

func (app App) GetCommon() (common PICSAppCommon, err error) {

	common = PICSAppCommon{}

	err = helpers.Unmarshal([]byte(app.Common), &common)
	log.Err(err)
	return common, err
}

func (app App) GetConfig() (config PICSAppConfig, err error) {

	config = PICSAppConfig{}

	err = helpers.Unmarshal([]byte(app.Config), &config)
	log.Err(err)
	return config, err
}

func (app App) GetDepots() (depots PicsDepots, err error) {

	err = helpers.Unmarshal([]byte(app.Depots), &depots)
	log.Err(err)
	return depots, err
}

func (app App) GetLaunch() (items []PICSAppConfigLaunchItem, err error) {

	err = helpers.Unmarshal([]byte(app.Launch), &items)
	log.Err(err)
	return items, err
}

func (app App) GetInstall() (install map[string]interface{}, err error) {

	install = map[string]interface{}{}

	err = helpers.Unmarshal([]byte(app.Install), &install)
	log.Err(err)
	return install, err
}

func (app App) GetLocalization() (localization map[string]interface{}, err error) {

	localization = map[string]interface{}{}

	err = helpers.Unmarshal([]byte(app.Localization), &localization)
	log.Err(err)
	return localization, err
}

func (app App) GetSystemRequirements() (systemRequirements map[string]interface{}, err error) {

	systemRequirements = map[string]interface{}{}

	err = helpers.Unmarshal([]byte(app.SystemRequirements), &systemRequirements)
	log.Err(err)
	return systemRequirements, err
}

func (app App) GetUFS() (ufs PICSAppUFS, err error) {

	ufs = PICSAppUFS{}

	err = helpers.Unmarshal([]byte(app.UFS), &ufs)
	log.Err(err)
	return ufs, err
}

func (app App) GetCommunityLink() string {
	name := config.Config.GameDBShortName.Get()
	return "https://steamcommunity.com/app/" + strconv.Itoa(app.ID) + "/?utm_source=" + name + "&utm_medium=link&utm_campaign=" + name
}

func (app App) GetStoreLink() string {
	name := config.Config.GameDBShortName.Get()
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

func (app App) GetScreenshots() (screenshots []AppImage, err error) {

	err = helpers.Unmarshal([]byte(app.Screenshots), &screenshots)
	log.Err(err)
	return screenshots, err
}

func (app App) GetMovies() (movies []AppVideo, err error) {

	err = helpers.Unmarshal([]byte(app.Movies), &movies)
	log.Err(err)
	return movies, err
}

func (app App) GetSteamSpy() (ss AppSteamSpy, err error) {

	err = helpers.Unmarshal([]byte(app.SteamSpy), &ss)
	log.Err(err)
	return ss, err
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

func (app App) GetAchievements() (achievements []AppAchievement, err error) {

	err = helpers.Unmarshal([]byte(app.Achievements), &achievements)
	log.Err(err)
	return achievements, err
}

func (app App) GetStats() (stats []AppStat, err error) {

	err = helpers.Unmarshal([]byte(app.Stats), &stats)
	log.Err(err)
	return stats, err
}

func (app App) GetPlatforms() (platforms []string, err error) {

	err = helpers.Unmarshal([]byte(app.Platforms), &platforms)
	log.Err(err)
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
	log.Err(err)
	return dlcs, err
}

func (app App) GetPackages() (packages []int, err error) {

	err = helpers.Unmarshal([]byte(app.Packages), &packages)
	log.Err(err)
	return packages, err
}

func (app App) GetReviews() (reviews AppReviewSummary, err error) {

	err = helpers.Unmarshal([]byte(app.Reviews), &reviews)
	log.Err(err)
	return reviews, err
}

func (app App) GetGenres() (genres []AppGenre, err error) {

	err = helpers.Unmarshal([]byte(app.Genres), &genres)
	log.Err(err)
	return genres, err
}

func (app App) GetCategories() (categories []string, err error) {

	err = helpers.Unmarshal([]byte(app.Categories), &categories)
	log.Err(err)
	return categories, err
}

func (app App) GetTagIDs() (tags []int, err error) {

	err = helpers.Unmarshal([]byte(app.StoreTags), &tags)
	log.Err(err)
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
	log.Err(err)
	return developers, err
}

func (app App) GetPublishers() (publishers []string, err error) {

	err = helpers.Unmarshal([]byte(app.Publishers), &publishers)
	log.Err(err)
	return publishers, err
}

func (app App) GetName() (name string) {
	return getAppName(app.ID, app.Name)
}

type SteamSpyAppResponse struct {
	Appid     int    `json:"appid"`
	Name      string `json:"name"`
	Developer string `json:"developer"`
	Publisher string `json:"publisher"`
	// ScoreRank      int    `json:"score_rank"` // Can be empty string
	Positive       int    `json:"positive"`
	Negative       int    `json:"negative"`
	Userscore      int    `json:"userscore"`
	Owners         string `json:"owners"`
	AverageForever int    `json:"average_forever"`
	Average2Weeks  int    `json:"average_2weeks"`
	MedianForever  int    `json:"median_forever"`
	Median2Weeks   int    `json:"median_2weeks"`
	Price          string `json:"price"`
	Initialprice   string `json:"initialprice"`
	Discount       string `json:"discount"`
	Languages      string `json:"languages"`
	Genre          string `json:"genre"`
	Ccu            int    `json:"ccu"`
	// Tags           map[string]int `json:"tags"` // Can be an empty slice
}

func (a SteamSpyAppResponse) GetOwners() (ret []int) {

	owners := strings.Replace(a.Owners, ",", "", -1)
	owners = strings.Replace(owners, " ", "", -1)
	ownersStrings := strings.Split(owners, "..")
	ownersInts := helpers.StringSliceToIntSlice(ownersStrings)
	if len(ownersInts) == 2 {
		return ownersInts
	}
	return ret
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

func GetApp(id int, columns []string) (app App, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return app, err
	}

	db = db.First(&app, id)
	if db.Error != nil {
		return app, db.Error
	}

	if len(columns) > 0 {
		db = db.Select(columns)
		if db.Error != nil {
			return app, db.Error
		}
	}

	if app.ID == 0 {
		return app, ErrRecordNotFound
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

func GetAppsWithPackages() (apps []App, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return apps, err
	}

	db = db.Select([]string{"packages"})
	db = db.Where("JSON_DEPTH(packages) = 2")

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

func GetAppPath(id int, name string) string {

	p := "/apps/" + strconv.Itoa(id)

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

func GetAppIcon(id int, icon string) string {

	if icon == "" {
		return DefaultAppIcon
	} else if strings.HasPrefix(icon, "/") || strings.HasPrefix(icon, "http") {
		return icon
	}
	return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(id) + "/" + icon + ".jpg"
}

type AppImage struct {
	PathFull      string `json:"f"`
	PathThumbnail string `json:"t"`
}

type AppVideo struct {
	PathFull      string `json:"f"`
	PathThumbnail string `json:"s"`
	Title         string `json:"t"`
}

type AppAchievement struct {
	Name        string  `json:"n"`
	Icon        string  `json:"i"`
	Description string  `json:"d"`
	Completed   float64 `json:"c"`
}

type AppStat struct {
	Name        string `json:"n"`
	Default     int    `json:"d"`
	DisplayName string `json:"o"`
}

type AppSteamSpy struct {
	SSAveragePlaytimeTwoWeeks int `json:"aw"`
	SSAveragePlaytimeForever  int `json:"af"`
	SSMedianPlaytimeTwoWeeks  int `json:"mw"`
	SSMedianPlaytimeForever   int `json:"mf"`
	SSOwnersLow               int `json:"ol"`
	SSOwnersHigh              int `json:"oh"`
}

type AppReviewSummary struct {
	Positive int
	Negative int
	Reviews  []AppReview
}

func (r AppReviewSummary) GetTotal() int {
	return r.Negative + r.Positive
}

func (r AppReviewSummary) GetPositivePercent() float64 {
	return float64(r.Positive) / float64(r.GetTotal()) * 100
}

func (r AppReviewSummary) GetNegativePercent() float64 {
	return float64(r.Negative) / float64(r.GetTotal()) * 100
}

type AppReview struct {
	Review     template.HTML `json:"r"`
	Vote       bool          `json:"v"`
	VotesGood  int           `json:"g"`
	VotesFunny int           `json:"f"`
	Created    string        `json:"c"`
	PlayerPath string        `json:"p"`
	PlayerName string        `json:"n"`
}

type AppGenre struct {
	ID   int
	Name string
}
