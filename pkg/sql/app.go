package sql

import (
	"encoding/json"
	"errors"
	"html/template"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/Jleagle/steam-go/steam"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql/pics"
	"github.com/golang/snappy"
	"github.com/jinzhu/gorm"
)

const (
	platformWindows = "windows"
	platformMac     = "macos"
	platformLinux   = "linux"
)

var (
	ErrInvalidAppID = errors.New("invalid id")
)

type App struct {
	Achievements                  string    `gorm:"not null;column:achievements;type:text"`           // []AppAchievement
	AchievementsAverageCompletion float64   `gorm:"not null;column:achievements_average_completion"`  //
	AchievementsCount             int       `gorm:"not null;column:achievements_count"`               //
	AlbumMetaData                 string    `gorm:"not null;column:albummetadata;type:text"`          // AlbumMetaData
	Background                    string    `gorm:"not null;column:background"`                       //
	BundleIDs                     string    `gorm:"not null;column:bundle_ids"`                       // []int
	Categories                    string    `gorm:"not null;column:categories;type:json"`             // []int8
	ChangeNumber                  int       `gorm:"not null;column:change_number"`                    //
	ChangeNumberDate              time.Time `gorm:"not null;column:change_number_date;type:datetime"` //
	ClientIcon                    string    `gorm:"not null;column:client_icon"`                      //
	ComingSoon                    bool      `gorm:"not null;column:coming_soon"`                      //
	Common                        string    `gorm:"not null;column:common"`                           // PICSAppCommon
	Config                        string    `gorm:"not null;column:config"`                           // PICSAppConfig
	CreatedAt                     time.Time `gorm:"not null;column:created_at;type:datetime"`         //
	DemoIDs                       string    `gorm:"not null;column:demo_ids;type:json"`               // []int
	Depots                        string    `gorm:"not null;column:depots"`                           // Depots
	Developers                    string    `gorm:"not null;column:developers;type:json"`             // []int
	DLC                           string    `gorm:"not null;column:dlc;type:json"`                    // []int
	DLCCount                      int       `gorm:"not null;column:dlc_count"`                        //
	Extended                      string    `gorm:"not null;column:extended"`                         // PICSExtended
	GameID                        int       `gorm:"not null;column:game_id"`                          //
	GameName                      string    `gorm:"not null;column:game_name"`                        //
	Genres                        string    `gorm:"not null;column:genres;type:json"`                 // []int
	GroupID                       string    `gorm:"not null;column:group_id;type:varchar"`            //
	GroupFollowers                int       `gorm:"not null;column:group_followers;type:int"`         //
	Homepage                      string    `gorm:"not null;column:homepage"`                         //
	Icon                          string    `gorm:"not null;column:icon"`                             //
	ID                            int       `gorm:"not null;column:id;primary_key"`                   //
	Install                       string    `gorm:"not null;column:install"`                          // map[string]interface{}
	IsFree                        bool      `gorm:"not null;column:is_free;type:tinyint(1)"`          //
	Items                         int       `gorm:"not null;column:items;type:int"`                   //
	ItemsDigest                   string    `gorm:"not null;column:items_digest"`                     //
	Launch                        string    `gorm:"not null;column:launch"`                           // []db.PICSAppConfigLaunchItem
	Localization                  string    `gorm:"not null;column:localization"`                     // map[string]interface{}
	Logo                          string    `gorm:"not null;column:logo"`                             //
	MetacriticScore               int8      `gorm:"not null;column:metacritic_score"`                 //
	MetacriticURL                 string    `gorm:"not null;column:metacritic_url"`                   //
	Movies                        string    `gorm:"not null;column:movies;type:text"`                 // []AppVideo
	Name                          string    `gorm:"not null;column:name"`                             //
	NewsIDs                       string    `gorm:"not null;column:news_ids"`                         // []int64
	Packages                      string    `gorm:"not null;column:packages;type:json"`               // []int
	Platforms                     string    `gorm:"not null;column:platforms;type:json"`              // []string
	PlayerAverageWeek             float64   `gorm:"not null;column:player_avg_week;type:float"`       //
	PlayerPeakAllTime             int       `gorm:"not null;column:player_peak_alltime"`              //
	PlayerPeakWeek                int       `gorm:"not null;column:player_peak_week"`                 //
	PlayerTrend                   int64     `gorm:"not null;column:player_trend"`                     //
	PlaytimeAverage               float64   `gorm:"not null;column:playtime_average"`                 // Minutes
	PlaytimeTotal                 int64     `gorm:"not null;column:playtime_total"`                   // Minutes
	Prices                        string    `gorm:"not null;column:prices"`                           // ProductPrices
	PublicOnly                    bool      `gorm:"not null;column:public_only"`                      //
	Publishers                    string    `gorm:"not null;column:publishers;type:json"`             // []int
	RelatedAppIDs                 string    `gorm:"not null;column:related_app_ids;type:json"`        // []int
	RelatedOwnersAppIDs           string    `gorm:"not null;column:related_owners_app_ids;type:json"` // []int
	ReleaseDate                   string    `gorm:"not null;column:release_date"`                     //
	ReleaseDateUnix               int64     `gorm:"not null;column:release_date_unix"`                //
	ReleaseState                  string    `gorm:"not null;column:release_state"`                    //
	Reviews                       string    `gorm:"not null;column:reviews"`                          // AppReviewSummary
	ReviewsScore                  float64   `gorm:"not null;column:reviews_score"`                    //
	Screenshots                   string    `gorm:"not null;column:screenshots;type:text"`            // []AppImage
	ShortDescription              string    `gorm:"not null;column:description_short"`                //
	Stats                         string    `gorm:"not null;column:stats;type:text"`                  // []AppStat
	SteamSpy                      string    `gorm:"not null;column:steam_spy"`                        // AppSteamSpy
	SystemRequirements            string    `gorm:"not null;column:system_requirements"`              // map[string]interface{}
	Tags                          string    `gorm:"not null;column:tags;type:json"`                   // []int
	TwitchID                      int       `gorm:"not null;column:twitch_id"`                        //
	TwitchURL                     string    `gorm:"not null;column:twitch_url"`                       //
	Type                          string    `gorm:"not null;column:type"`                             //
	UFS                           string    `gorm:"not null;column:ufs"`                              // PICSAppUFS
	UpdatedAt                     time.Time `gorm:"not null;column:updated_at;type:datetime"`         //
	Version                       string    `gorm:"not null;column:version"`                          //
	WishlistAvgPosition           float64   `gorm:"not null;column:wishlist_avg_position"`            //
	WishlistCount                 int       `gorm:"not null;column:wishlist_count"`                   //
}

func (app *App) BeforeCreate(scope *gorm.Scope) error {
	return app.Before(scope)
}

func (app *App) BeforeSave(scope *gorm.Scope) error {

	err := app.Before(scope)
	if err != nil {
		return err
	}

	return app.SaveToMongo()
}

func (app *App) Before(scope *gorm.Scope) error {

	if app.Achievements == "" {
		app.Achievements = "[]"
	}
	if app.BundleIDs == "" || app.BundleIDs == "null" {
		app.BundleIDs = "[]"
	}
	if app.Categories == "" {
		app.Categories = "[]"
	}
	if app.ChangeNumberDate.IsZero() {
		app.ChangeNumberDate = time.Now()
	}
	if app.Common == "" {
		app.Common = "{}"
	}
	if app.Config == "" {
		app.Config = "{}"
	}
	if app.Depots == "" {
		app.Depots = "{}"
	}
	if app.Developers == "" {
		app.Developers = "[]"
	}
	if app.DemoIDs == "" {
		app.DemoIDs = "[]"
	}
	if app.DLC == "" {
		app.DLC = "[]"
	}
	if app.Extended == "" {
		app.Extended = "{}"
	}
	if app.Genres == "" {
		app.Genres = "[]"
	}
	if app.Install == "" {
		app.Install = "{}"
	}
	if app.Launch == "" {
		app.Launch = "[]"
	}
	if app.Localization == "" {
		app.Localization = "{}"
	}
	if app.Movies == "" {
		app.Movies = "[]"
	}
	if app.NewsIDs == "" {
		app.NewsIDs = "[]"
	}
	if app.Packages == "" {
		app.Packages = "[]"
	}
	if app.Platforms == "" {
		app.Platforms = "[]"
	}
	if app.Prices == "" {
		app.Prices = "{}"
	}
	if app.Publishers == "" {
		app.Publishers = "[]"
	}
	if app.RelatedAppIDs == "" {
		app.RelatedAppIDs = "[]"
	}
	if app.RelatedOwnersAppIDs == "" {
		app.RelatedOwnersAppIDs = "[]"
	}
	if app.Reviews == "" {
		app.Reviews = "{}"
	}
	if app.Stats == "" {
		app.Stats = "[]"
	}
	if app.Screenshots == "" {
		app.Screenshots = "[]"
	}
	if app.SteamSpy == "" {
		app.SteamSpy = "{}"
	}
	if app.SystemRequirements == "" {
		app.SystemRequirements = "{}"
	}
	if app.Tags == "" {
		app.Tags = "[]"
	}
	if app.UFS == "" {
		app.UFS = "{}"
	}

	return nil
}

// Sync sql apps to mongo apps
func (app App) SaveToMongo() error {

	mApp := mongo.App{}
	mApp.Achievements = app.Achievements
	mApp.AchievementsAverageCompletion = app.AchievementsAverageCompletion
	mApp.AchievementsCount = app.AchievementsCount
	mApp.AlbumMetaData = app.AlbumMetaData
	mApp.Background = app.Background
	mApp.BundleIDs = app.getBundleIDs()
	mApp.Categories = app.GetCategoryIDs()
	mApp.ChangeNumber = app.ChangeNumber
	mApp.ChangeNumberDate = app.ChangeNumberDate
	mApp.ClientIcon = app.ClientIcon
	mApp.ComingSoon = app.ComingSoon
	mApp.Common = app.GetCommon().Map()
	mApp.Config = app.GetConfig().Map()
	mApp.CreatedAt = app.CreatedAt
	mApp.DemoIDs = app.GetDemoIDs()
	mApp.Depots = app.Depots
	mApp.Developers = app.GetDeveloperIDs()
	mApp.DLC, _ = app.GetDLCIDs()
	mApp.DLCCount = app.DLCCount
	mApp.Extended = app.GetExtended().Map()
	mApp.GameID = app.GameID
	mApp.GameName = app.GameName
	mApp.Genres = app.GetGenreIDs()
	mApp.GroupID = app.GroupID
	mApp.GroupFollowers = app.GroupFollowers
	mApp.Homepage = app.Homepage
	mApp.Icon = app.Icon
	mApp.ID = app.ID
	mApp.Install = app.Install
	mApp.IsFree = app.IsFree
	mApp.Items = app.Items
	mApp.ItemsDigest = app.ItemsDigest
	mApp.Launch = app.Launch
	mApp.Localization = app.Localization
	mApp.Logo = app.Logo
	mApp.MetacriticScore = app.MetacriticScore
	mApp.MetacriticURL = app.MetacriticURL
	mApp.Movies = app.Movies
	mApp.Name = app.Name
	mApp.NewsIDs = app.GetNewsIDs()
	mApp.Packages = app.GetPackageIDs()
	mApp.Platforms = app.GetPlatforms()
	mApp.PlayerAverageWeek = app.PlayerAverageWeek
	mApp.PlayerPeakAllTime = app.PlayerPeakAllTime
	mApp.PlayerPeakWeek = app.PlayerPeakWeek
	mApp.PlayerTrend = app.PlayerTrend
	mApp.PlaytimeAverage = app.PlaytimeAverage
	mApp.PlaytimeTotal = app.PlaytimeTotal
	mApp.Prices = app.Prices
	mApp.PublicOnly = app.PublicOnly
	mApp.Publishers = app.GetPublisherIDs()
	mApp.RelatedAppIDs, _ = app.GetRelatedIDs()
	mApp.RelatedOwnersAppIDs, _ = app.GetRelatedOwnerIDs()
	mApp.ReleaseDate = app.ReleaseDate
	mApp.ReleaseDateUnix = app.ReleaseDateUnix
	mApp.ReleaseState = app.ReleaseState
	mApp.Reviews = app.Reviews
	mApp.ReviewsScore = app.ReviewsScore
	mApp.Screenshots = app.Screenshots
	mApp.ShortDescription = app.ShortDescription
	mApp.Stats = app.Stats
	mApp.SteamSpy = app.SteamSpy
	mApp.SystemRequirements = app.SystemRequirements
	mApp.Tags = app.GetTagIDs()
	mApp.TwitchID = app.TwitchID
	mApp.TwitchURL = app.TwitchURL
	mApp.Type = app.Type
	mApp.UFS = app.GetUFS().Map()
	mApp.UpdatedAt = app.UpdatedAt
	mApp.Version = app.Version
	mApp.WishlistAvgPosition = app.WishlistAvgPosition
	mApp.WishlistCount = app.WishlistCount

	return mApp.Save()
}

func (app App) GetID() int {
	return app.ID
}

func (app App) GetProductType() helpers.ProductType {
	return helpers.ProductTypeApp
}

func (app App) GetPath() string {
	return helpers.GetAppPath(app.ID, app.Name)
}

func (app App) GetType() (ret string) {
	return helpers.GetAppType(app.Type)
}

func (app App) GetReviewScore() string {

	return helpers.FloatToString(app.ReviewsScore, 2) + "%"
}

func (app App) GetDaysToRelease() string {

	return helpers.GetDaysToRelease(app.ReleaseDateUnix)
}

func (app App) GetReleaseState() (ret string) {
	return helpers.GetAppReleaseState(app.ReleaseState)
}

func (app App) GetReleaseDateNice() string {
	return helpers.GetAppReleaseDateNice(app.ReleaseDateUnix, app.ReleaseDate)
}

func (app App) GetUpdatedNice() string {
	return app.UpdatedAt.Format(helpers.DateYearTime)
}

func (app App) GetPICSUpdatedNice() string {

	d := app.ChangeNumberDate

	// 0000-01-01 00:00:00
	if d.Unix() == -62167219200 {
		return "-"
	}

	if d.IsZero() {
		return "-"
	}
	return d.Format(helpers.DateYearTime)
}

func (app App) GetIcon() (ret string) {
	return helpers.GetAppIcon(app.ID, app.Icon)
}

func (app App) GetFollowers() (ret string) {

	if app.GroupID == "" {
		return "-"
	}

	return humanize.Comma(int64(app.GroupFollowers))
}

func (app App) GetPrices() (prices ProductPrices) {

	prices = ProductPrices{}

	err := helpers.Unmarshal([]byte(app.Prices), &prices)
	log.Err(err)

	return prices
}

func (app App) GetPrice(code steam.ProductCC) (price ProductPrice) {

	return app.GetPrices().Get(code)
}

func (app App) GetNewsIDs() (ids []int64) {

	if app.NewsIDs == "" {
		return ids
	}

	err := helpers.Unmarshal([]byte(app.NewsIDs), &ids)
	log.Err(err)

	return ids
}

func (app App) GetExtended() (extended pics.PICSKeyValues) {

	extended = pics.PICSKeyValues{}

	err := helpers.Unmarshal([]byte(app.Extended), &extended)
	log.Err(err)

	return extended
}

func (app App) GetAlbum() (data pics.AlbumMetaData) {

	if len(app.AlbumMetaData) < 3 {
		return data
	}

	data = pics.AlbumMetaData{}
	err := helpers.Unmarshal([]byte(app.AlbumMetaData), &data)
	log.Err(err)

	return data
}

func (app App) GetCommon() (common pics.PICSKeyValues) {

	common = pics.PICSKeyValues{}
	err := helpers.Unmarshal([]byte(app.Common), &common)
	log.Err(err)

	return common
}

func (app App) GetConfig() (config pics.PICSKeyValues) {

	config = pics.PICSKeyValues{}
	err := helpers.Unmarshal([]byte(app.Config), &config)
	log.Err(err)

	return config
}

func (app App) GetUFS() (ufs pics.PICSKeyValues) {

	ufs = pics.PICSKeyValues{}
	err := helpers.Unmarshal([]byte(app.UFS), &ufs)
	log.Err(err)

	return ufs
}

func (app App) GetDepots() (depots pics.Depots, err error) {

	err = helpers.Unmarshal([]byte(app.Depots), &depots)
	log.Err(err)
	return depots, err
}

func (app App) GetLaunch() (items []pics.PICSAppConfigLaunchItem, err error) {

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

func (app App) GetLocalization() (localization pics.Localisation, err error) {

	if app.Localization == "" || app.Localization == "{}" {
		return localization, nil
	}

	decoded, err := snappy.Decode(nil, []byte(app.Localization))
	if err == nil {
		app.Localization = string(decoded)
	}

	localization = pics.Localisation{}
	err = helpers.Unmarshal([]byte(app.Localization), &localization)
	return localization, err
}

func (app *App) SetLocalization(localization pics.Localisation) {

	b, err := json.Marshal(localization)
	if err != nil {
		log.Err(app.ID, err)
		app.Localization = ""
		return
	}

	if len(b) == 0 || string(b) == "{}" {
		app.Localization = ""
		return
	}

	// Snappy to save space
	encoded := snappy.Encode(nil, b)
	app.Localization = string(encoded)
}

func (app App) GetSystemRequirements() (ret []SystemRequirement, err error) {

	systemRequirements := map[string]interface{}{}

	err = helpers.Unmarshal([]byte(app.SystemRequirements), &systemRequirements)
	log.Err(err)

	flattened := helpers.FlattenMap(systemRequirements)

	for k, v := range flattened {
		if val, ok := v.(string); ok {
			ret = append(ret, SystemRequirement{Key: k, Val: val})
		}
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Key < ret[j].Key
	})

	return ret, err
}

type SystemRequirement struct {
	Key string
	Val string
}

func (sr SystemRequirement) Format() template.HTML {

	switch sr.Val {
	case "0":
		return `<i class="fas fa-times text-danger"></i>`
	case "1":
		return `<i class="fas fa-check text-success"></i>`
	case "warn":
		return `<span class="text-warning">Warn</span>`
	case "deny":
		return `<span class="text-danger">Deny</span>`
	default:
		return template.HTML(sr.Val)
	}
}

func (app App) IsOnSale() bool {

	common := app.GetCommon()

	if common.GetValue("app_retired_publisher_request") == "1" {
		return false
	}

	return true
}

func (app App) GetOnlinePlayers() (players int64, err error) {

	var item = memcache.MemcacheAppPlayersRow(app.ID)

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &players, func() (interface{}, error) {

		builder := influxql.NewBuilder()
		builder.AddSelect("player_count", "")
		builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
		builder.AddWhere("app_id", "=", app.ID)
		builder.AddOrderBy("time", false)
		builder.SetLimit(1)

		resp, err := influx.InfluxQuery(builder.String())

		return influx.GetFirstInfluxInt(resp), err
	})

	return players, err
}

func (app App) GetCommunityLink() string {
	name := config.Config.GameDBShortName.Get()
	return "https://steamcommunity.com/app/" + strconv.Itoa(app.ID) + "?utm_source=" + name + "&utm_medium=link&curator_clanid=" // todo curator_clanid
}

func (app App) GetStoreLink() string {
	return helpers.GetAppStoreLink(app.ID)
}

func (app App) GetPCGamingWikiLink() string {
	return "https://pcgamingwiki.com/api/appid.php?appid=" + strconv.Itoa(app.ID)
}

func (app App) GetHeaderImage() string {
	return "https://steamcdn-a.akamaihd.net/steam/apps/" + strconv.Itoa(app.ID) + "/header.jpg"
}

// func (app App) GetHeaderImage2() string {
//
// 	params := url.Values{}
// 	params.Set("url", app.GetHeaderImage())
// 	params.Set("q", "10")
//
// 	return "https://images.weserv.nl?" + params.Encode()
// }

func (app App) GetInstallLink() template.URL {
	return template.URL("steam://install/" + strconv.Itoa(app.ID))
}

func (app App) GetMetacriticLink() template.URL {
	return template.URL("https://www.metacritic.com/game/" + app.MetacriticURL)
}

func (app App) GetScreenshots() (screenshots []AppImage) {

	err := helpers.Unmarshal([]byte(app.Screenshots), &screenshots)
	log.Err(err)

	return screenshots
}

func (app App) GetMovies() (movies []AppVideo) {

	err := helpers.Unmarshal([]byte(app.Movies), &movies)
	log.Err(err)

	return movies
}

func (app App) GetSteamSpy() (ss AppSteamSpy) {

	err := helpers.Unmarshal([]byte(app.SteamSpy), &ss)
	log.Err(err)

	return ss
}

func (app App) GetCoopTags() (string, error) {

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
	for _, v := range app.GetTagIDs() {
		if val, ok := tagMap[v]; ok {
			coopTags = append(coopTags, val)
		}
	}

	return strings.Join(coopTags, ", "), nil
}

// Template
func (app App) GetAchievements() (achievements []AppAchievement) {

	err := helpers.Unmarshal([]byte(app.Achievements), &achievements)
	log.Err(err)

	return achievements
}

func (app App) GetStats() (stats []AppStat) {

	err := helpers.Unmarshal([]byte(app.Stats), &stats)
	log.Err(err)

	return stats
}

func (app App) GetDemoIDs() (demos []int) {

	err := helpers.Unmarshal([]byte(app.DemoIDs), &demos)
	log.Err(err)

	return demos
}

func (app App) GetDemos() (demos []App, err error) {

	demos = []App{} // Needed for marshalling into type

	var item = memcache.MemcacheAppDemos(app.ID)

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &demos, func() (interface{}, error) {
		return GetAppsByID(app.GetDemoIDs(), []string{"id", "name"})
	})

	return demos, err
}

func (app App) GetPlatforms() (platforms []string) {

	platforms = []string{} // Needed for marshalling into array

	err := helpers.Unmarshal([]byte(app.Platforms), &platforms)
	log.Err(err)

	return platforms
}

func (app App) GetPlatformImages() (ret template.HTML, err error) {

	platforms := app.GetPlatforms()

	if len(platforms) == 0 {
		return "", nil
	}

	if helpers.SliceHasString(platforms, platformWindows) {
		ret = ret + `<a href="/apps?platforms=windows"><i class="fab fa-windows" data-toggle="tooltip" data-placement="top" title="Windows"></i></a>`
	} else {
		ret = ret + `<span class="space"></span>`
	}

	if helpers.SliceHasString(platforms, platformMac) {
		ret = ret + `<a href="/apps?platforms=macos"><i class="fab fa-apple" data-toggle="tooltip" data-placement="top" title="Mac"></i></a>`
	} else {
		ret = ret + `<span class="space"></span>`
	}

	if helpers.SliceHasString(platforms, platformLinux) {
		ret = ret + `<a href="/apps?platforms=linux"><i class="fab fa-linux" data-toggle="tooltip" data-placement="top" title="Linux"></i></a>`
	} else {
		ret = ret + `<span class="space"></span>`
	}

	return ret, nil
}

func (app App) GetDLCIDs() (dlcs []int, err error) {

	err = helpers.Unmarshal([]byte(app.DLC), &dlcs)
	log.Err(err)

	if len(dlcs) == 0 {
		dlcs = []int{} // Needed for marshalling into type
	}

	return dlcs, err
}

func (app App) GetDLCs() (apps []App, err error) {

	var item = memcache.MemcacheAppDLC(app.ID)

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &apps, func() (interface{}, error) {

		ids, err := app.GetDLCIDs()
		if err != nil {
			return apps, err
		}

		return GetAppsByID(ids, []string{"id", "name"})
	})

	if len(apps) == 0 {
		apps = []App{} // Needed for marshalling into type
	}

	return apps, err
}

func (app App) GetPackageIDs() (packages []int) {

	packages = []int{} // Needed for marshalling into type

	err := helpers.Unmarshal([]byte(app.Packages), &packages)
	log.Err(err)

	return packages
}

func (app App) GetReviews() (reviews AppReviewSummary) {

	reviews = AppReviewSummary{} // Needed for marshalling into type

	err := helpers.Unmarshal([]byte(app.Reviews), &reviews)
	log.Err(err)

	return reviews
}

func (app App) GetGenreIDs() (genres []int) {

	genres = []int{}

	err := helpers.Unmarshal([]byte(app.Genres), &genres)
	log.Err(err)

	return genres
}

func (app App) GetRelatedIDs() (apps []int, err error) {

	apps = []int{}

	err = helpers.Unmarshal([]byte(app.RelatedAppIDs), &apps)
	return apps, err
}

func (app App) GetRelatedOwnerIDs() (apps []helpers.TupleInt, err error) {

	err = helpers.Unmarshal([]byte(app.RelatedOwnersAppIDs), &apps)
	return apps, err
}

func (app App) GetRelatedApps() (apps []App, err error) {

	apps = []App{} // Needed for marshalling into type

	var item = memcache.MemcacheAppRelated(app.ID)

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &apps, func() (interface{}, error) {

		ids, err := app.GetRelatedIDs()
		if err != nil {
			return apps, err
		}

		return GetAppsByID(ids, []string{"id", "name"})
	})

	return apps, err
}

func (app App) GetGenres() (genres []Genre, err error) {

	genres = []Genre{} // Needed for marshalling into type

	var item = memcache.MemcacheAppGenres(app.ID)

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &genres, func() (interface{}, error) {
		return GetGenresByID(app.GetGenreIDs(), []string{"id", "name"})
	})

	return genres, err
}

func (app App) GetCategoryIDs() (categories []int) {

	categories = []int{} // Needed for marshalling into array

	err := helpers.Unmarshal([]byte(app.Categories), &categories)
	log.Err(err)

	return categories
}

func (app App) GetCategories() (categories []Category, err error) {

	var item = memcache.MemcacheAppCategories(app.ID)

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &categories, func() (interface{}, error) {

		return GetCategoriesByID(app.GetCategoryIDs(), []string{"id", "name"})
	})

	if len(categories) == 0 {
		categories = []Category{} // Needed for marshalling into type
	}

	return categories, err
}

func (app App) GetTagIDs() (tags []int) {

	tags = []int{} // Needed for marshalling into type

	if app.Tags == "" || app.Tags == "null" || app.Tags == "[]" {
		return tags
	}

	err := helpers.Unmarshal([]byte(app.Tags), &tags)
	if err != nil {
		log.Err(err)
		return
	}

	log.Err(err)

	return tags
}

func (app App) GetTags() (tags []Tag, err error) {

	var item = memcache.MemcacheAppTags(app.ID)

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &tags, func() (interface{}, error) {

		return GetTagsByID(app.GetTagIDs(), []string{"id", "name"})
	})

	if len(tags) == 0 {
		tags = []Tag{} // Needed for marshalling into type
	}

	return tags, err
}

func (app App) GetDeveloperIDs() (developers []int) {

	developers = []int{}

	err := helpers.Unmarshal([]byte(app.Developers), &developers)
	log.Err(err)

	return developers
}

func (app App) GetDevelopers() (developers []Developer, err error) {

	var item = memcache.MemcacheAppDevelopers(app.ID)

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &developers, func() (interface{}, error) {
		return GetDevelopersByID(app.GetDeveloperIDs(), []string{"id", "name"})
	})

	if len(developers) == 0 {
		developers = []Developer{} // Needed for marshalling into type
	}

	return developers, err
}

func (app App) GetPublisherIDs() (publishers []int) {

	publishers = []int{} // Needed for marshalling into type

	err := helpers.Unmarshal([]byte(app.Publishers), &publishers)
	log.Err(err)

	return publishers
}

func (app App) GetPublishers() (publishers []Publisher, err error) {

	publishers = []Publisher{} // Needed for marshalling into type

	var item = memcache.MemcacheAppPublishers(app.ID)

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &publishers, func() (interface{}, error) {
		return GetPublishersByID(app.GetPublisherIDs(), []string{"id", "name"})
	})

	return publishers, err
}

func (app App) getBundleIDs() (ids []int) {

	ids = []int{} // Needed for marshalling into type

	err := helpers.Unmarshal([]byte(app.BundleIDs), &ids)
	log.Err(err)

	return ids
}

func (app App) GetBundles() (bundles []Bundle, err error) {

	var item = memcache.MemcacheAppBundles(app.ID)

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &bundles, func() (interface{}, error) {
		return GetBundlesByID(app.getBundleIDs(), []string{})
	})

	if len(bundles) == 0 {
		bundles = []Bundle{} // Needed for marshalling into type
	}

	return bundles, err
}

func (app App) GetName() string {
	return helpers.GetAppName(app.ID, app.Name)
}

func (app App) GetMetaImage() string {

	ss := app.GetScreenshots()
	if len(ss) == 0 {
		return app.GetHeaderImage()
	}
	return ss[0].PathFull
}

func PopularApps() (apps []App, err error) {

	var item = memcache.MemcachePopularApps

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &apps, func() (interface{}, error) {

		db, err := GetMySQLClient()
		if err != nil {
			return apps, err
		}

		db = db.Select([]string{"id", "name", "player_peak_week", "background"})
		db = db.Where("type = ?", "game")
		db = db.Order("player_peak_week desc")
		db = db.Limit(30)
		db = db.Find(&apps)

		return apps, err
	})

	return apps, err
}

func PopularNewApps() (apps []App, err error) {

	var item = memcache.MemcachePopularNewApps

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &apps, func() (interface{}, error) {

		db, err := GetMySQLClient()
		if err != nil {
			return apps, err
		}

		db = db.Select([]string{"id", "name", "player_peak_week"})
		db = db.Where("type = ?", "game")
		db = db.Where("release_date_unix > ?", time.Now().AddDate(0, 0, -config.Config.NewReleaseDays.GetInt()).Unix())
		db = db.Order("player_peak_week desc")
		db = db.Limit(25)
		db = db.Find(&apps)

		return apps, err
	})

	return apps, err
}

func TrendingApps() (apps []App, err error) {

	var item = memcache.MemcacheTrendingApps

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &apps, func() (interface{}, error) {

		db, err := GetMySQLClient()
		if err != nil {
			return apps, err
		}

		db = db.Select([]string{"id", "name", "player_trend"})
		db = db.Order("player_trend desc")
		db = db.Limit(10)
		db = db.Find(&apps)

		return apps, err
	})

	return apps, err
}

func WishlistedApps() (appsMap map[int]bool, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return appsMap, err
	}

	var apps []App
	db = db.Select([]string{"id"})
	db = db.Where("wishlist_count > ?", 0)
	db = db.Find(&apps)

	appsMap = map[int]bool{}
	for _, app := range apps {
		appsMap[app.ID] = true
	}

	return appsMap, err
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

	owners := strings.ReplaceAll(a.Owners, ",", "")
	owners = strings.ReplaceAll(owners, " ", "")
	ownersStrings := strings.Split(owners, "..")
	return helpers.StringSliceToIntSlice(ownersStrings)
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

	if id == 0 {
		id = 753
	}

	if !helpers.IsValidAppID(id) {
		return app, ErrInvalidAppID
	}

	db, err := GetMySQLClient()
	if err != nil {
		return app, err
	}

	if len(columns) > 0 {
		db = db.Select(columns)
		if db.Error != nil {
			return app, db.Error
		}
	}

	db = db.First(&app, id)
	if db.Error != nil {
		return app, db.Error
	}

	if app.ID == 0 {
		return app, ErrRecordNotFound
	}

	return app, nil
}

func SearchApp(s string, columns []string) (app App, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return app, err
	}

	if len(columns) > 0 {
		db = db.Select(columns)
		if db.Error != nil {
			return app, db.Error
		}
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return app, err
	}

	if helpers.IsValidAppID(i) {
		db = db.First(&app, "id = ?", s)
	} else {
		db = db.First(&app, "name = ?", s)
	}

	if db.Error != nil {
		return app, db.Error
	}

	if app.ID == 0 {
		return app, ErrRecordNotFound
	}

	return app, nil
}

func GetAppsByID(ids []int, columns []string) (apps []App, err error) {

	if len(ids) == 0 {
		return apps, nil
	}

	db, err := GetMySQLClient()
	if err != nil {
		return apps, err
	}

	ids = helpers.Unique(ids)

	chunks := helpers.ChunkInts(ids, 100)
	for _, chunk := range chunks {

		db = db.New()

		if len(columns) > 0 {
			db = db.Select(columns)
		}

		var appsChunk []App
		db = db.Where("id IN (?)", chunk).Find(&appsChunk)
		if db.Error != nil {
			log.Err(db.Error)
			return apps, db.Error
		}

		apps = append(apps, appsChunk...)
	}

	return apps, nil
}

func GetAppsWithColumnDepth(column string, depth int, columns []string) (apps []App, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return apps, err
	}

	db = db.Select(columns)
	db = db.Where("JSON_DEPTH("+column+") = ?", depth)
	db = db.Order("id asc")

	db = db.Find(&apps)
	if db.Error != nil {
		return apps, db.Error
	}

	return apps, nil

}

func CountApps() (count int, err error) {

	var item = memcache.MemcacheAppsCount

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		var count int

		db, err := GetMySQLClient()
		if err != nil {
			return count, err
		}

		db.Model(&App{}).Count(&count)

		return count, db.Error
	})

	return count, err
}

func CountAppsWithAchievements() (count int, err error) {

	var item = memcache.MemcacheAppsWithAchievementsCount

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		var count int

		db, err := GetMySQLClient()
		if err != nil {
			return count, err
		}

		db.Model(&App{}).Where("achievements_count > 0").Count(&count)

		return count, db.Error
	})

	return count, err
}

//
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
	Active      bool    `json:"a"`
}

func (a AppAchievement) GetIcon() string {
	if strings.HasSuffix(a.Icon, ".jpg") {
		return a.Icon
	}
	return helpers.DefaultAppIcon
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

func (ss AppSteamSpy) GetSSAveragePlaytimeTwoWeeks() float64 {
	return helpers.RoundFloatTo1DP(float64(ss.SSAveragePlaytimeTwoWeeks) / 60)
}

func (ss AppSteamSpy) GetSSAveragePlaytimeForever() float64 {
	return helpers.RoundFloatTo1DP(float64(ss.SSAveragePlaytimeForever) / 60)
}

func (ss AppSteamSpy) GetSSMedianPlaytimeTwoWeeks() float64 {
	return helpers.RoundFloatTo1DP(float64(ss.SSMedianPlaytimeTwoWeeks) / 60)
}

func (ss AppSteamSpy) GetSSMedianPlaytimeForever() float64 {
	return helpers.RoundFloatTo1DP(float64(ss.SSMedianPlaytimeForever) / 60)
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
	Review     string `json:"r"`
	Vote       bool   `json:"v"`
	VotesGood  int    `json:"g"`
	VotesFunny int    `json:"f"`
	Created    string `json:"c"`
	PlayerPath string `json:"p"`
	PlayerName string `json:"n"`
}

func (ar AppReview) HTML() template.HTML {
	return template.HTML(ar.Review)
}
