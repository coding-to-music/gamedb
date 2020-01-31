package sql

import (
	"errors"
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
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

func (app App) GetIcon() (ret string) {
	return helpers.GetAppIcon(app.ID, app.Icon)
}

func (app App) GetPrices() (prices helpers.ProductPrices) {

	prices = helpers.ProductPrices{}

	err := helpers.Unmarshal([]byte(app.Prices), &prices)
	if err != nil {
		log.Err(err)
	}

	return prices
}

func (app App) GetHeaderImage() string {
	return "https://steamcdn-a.akamaihd.net/steam/apps/" + strconv.Itoa(app.ID) + "/header.jpg"
}

func (app App) GetScreenshots() (screenshots []helpers.AppImage) {

	err := helpers.Unmarshal([]byte(app.Screenshots), &screenshots)
	if err != nil {
		log.Err(err)
	}

	return screenshots
}

// Template
func (app App) GetPlatforms() (platforms []string) {

	platforms = []string{} // Needed for marshalling into array

	err := helpers.Unmarshal([]byte(app.Platforms), &platforms)
	if err != nil {
		log.Err(err)
	}

	return platforms
}

func (app App) GetReviews() (reviews helpers.AppReviewSummary) {

	reviews = helpers.AppReviewSummary{} // Needed for marshalling into type

	err := helpers.Unmarshal([]byte(app.Reviews), &reviews)
	if err != nil {
		log.Err(err)
	}

	return reviews
}

func (app App) GetGenreIDs() (genres []int) {

	genres = []int{}

	err := helpers.Unmarshal([]byte(app.Genres), &genres)
	if err != nil {
		log.Err(err)
	}

	return genres
}

func (app App) GetCategoryIDs() (categories []int) {

	categories = []int{} // Needed for marshalling into array

	err := helpers.Unmarshal([]byte(app.Categories), &categories)
	if err != nil {
		log.Err(err)
	}

	return categories
}

func (app App) GetTagIDs() (tags []int) {

	tags = []int{} // Needed for marshalling into type

	if app.Tags == "" || app.Tags == "null" || app.Tags == "[]" {
		return tags
	}

	err := helpers.Unmarshal([]byte(app.Tags), &tags)
	if err != nil {
		log.Err(err)
	}

	return tags
}

func (app App) GetDeveloperIDs() (developers []int) {

	developers = []int{}

	err := helpers.Unmarshal([]byte(app.Developers), &developers)
	if err != nil {
		log.Err(err)
	}

	return developers
}

func (app App) GetPublisherIDs() (publishers []int) {

	publishers = []int{} // Needed for marshalling into type

	err := helpers.Unmarshal([]byte(app.Publishers), &publishers)
	if err != nil {
		log.Err(err)
	}

	return publishers
}

func (app App) GetName() string {
	return helpers.GetAppName(app.ID, app.Name)
}
