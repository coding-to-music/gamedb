package mongo

import (
	"errors"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"go.mongodb.org/mongo-driver/bson"
)

var ErrInvalidAppID = errors.New("invalid app id")

type App struct {
	Achievements                  string    `gorm:"achievements"`                    // []AppAchievement
	AchievementsAverageCompletion float64   `gorm:"achievements_average_completion"` //
	AchievementsCount             int       `gorm:"achievements_count"`              //
	AlbumMetaData                 string    `gorm:"albummetadata"`                   // AlbumMetaData
	Background                    string    `gorm:"background"`                      //
	BundleIDs                     []int     `gorm:"bundle_ids"`                      //
	Categories                    []int     `gorm:"categories"`                      //
	ChangeNumber                  int       `gorm:"change_number"`                   //
	ChangeNumberDate              time.Time `gorm:"change_number_date"`              //
	ClientIcon                    string    `gorm:"client_icon"`                     //
	ComingSoon                    bool      `gorm:"coming_soon"`                     //
	Common                        string    `gorm:"common"`                          // PICSAppCommon
	Config                        string    `gorm:"config"`                          // PICSAppConfig
	CreatedAt                     time.Time `gorm:"created_at"`                      //
	DemoIDs                       []int     `gorm:"demo_ids"`                        //
	Depots                        string    `gorm:"depots"`                          // Depots
	Developers                    []int     `gorm:"developers"`                      //
	DLC                           []int     `gorm:"dlc"`                             //
	DLCCount                      int       `gorm:"dlc_count"`                       //
	Extended                      string    `gorm:"extended"`                        // PICSExtended
	GameID                        int       `gorm:"game_id"`                         //
	GameName                      string    `gorm:"game_name"`                       //
	Genres                        []int     `gorm:"genres"`                          //
	GroupID                       string    `gorm:"group_id"`                        //
	GroupFollowers                int       `gorm:"group_followers"`                 //
	Homepage                      string    `gorm:"homepage"`                        //
	Icon                          string    `gorm:"icon"`                            //
	ID                            int       `gorm:"id;primary_key"`                  //
	Install                       string    `gorm:"install"`                         // map[string]interface{}
	IsFree                        bool      `gorm:"is_free"`                         //
	Items                         int       `gorm:"items"`                           //
	ItemsDigest                   string    `gorm:"items_digest"`                    //
	Launch                        string    `gorm:"launch"`                          // []db.PICSAppConfigLaunchItem
	Localization                  string    `gorm:"localization"`                    // pics.Localisation
	Logo                          string    `gorm:"logo"`                            //
	MetacriticScore               int8      `gorm:"metacritic_score"`                //
	MetacriticURL                 string    `gorm:"metacritic_url"`                  //
	Movies                        string    `gorm:"movies"`                          // []AppVideo
	Name                          string    `gorm:"name"`                            //
	NewsIDs                       []int64   `gorm:"news_ids"`                        //
	Packages                      []int     `gorm:"packages"`                        //
	Platforms                     []string  `gorm:"platforms"`                       //
	PlayerAverageWeek             float64   `gorm:"player_avg_week"`                 //
	PlayerPeakAllTime             int       `gorm:"player_peak_alltime"`             //
	PlayerPeakWeek                int       `gorm:"player_peak_week"`                //
	PlayerTrend                   int64     `gorm:"player_trend"`                    //
	PlaytimeAverage               float64   `gorm:"playtime_average"`                // Minutes
	PlaytimeTotal                 int64     `gorm:"playtime_total"`                  // Minutes
	Prices                        string    `gorm:"prices"`                          // ProductPrices
	PublicOnly                    bool      `gorm:"public_only"`                     //
	Publishers                    []int     `gorm:"publishers"`                      //
	RelatedAppIDs                 []int     `gorm:"related_app_ids"`                 //
	ReleaseDate                   string    `gorm:"release_date"`                    //
	ReleaseDateUnix               int64     `gorm:"release_date_unix"`               //
	ReleaseState                  string    `gorm:"release_state"`                   //
	Reviews                       string    `gorm:"reviews"`                         // AppReviewSummary
	ReviewsScore                  float64   `gorm:"reviews_score"`                   //
	Screenshots                   string    `gorm:"screenshots"`                     // []AppImage
	ShortDescription              string    `gorm:"description_short"`               //
	Stats                         string    `gorm:"stats"`                           // []AppStat
	SteamSpy                      string    `gorm:"steam_spy"`                       // AppSteamSpy
	SystemRequirements            string    `gorm:"system_requirements"`             // map[string]interface{}
	Tags                          []int     `gorm:"tags"`                            //
	TwitchID                      int       `gorm:"twitch_id"`                       //
	TwitchURL                     string    `gorm:"twitch_url"`                      //
	Type                          string    `gorm:"type"`                            //
	UFS                           string    `gorm:"ufs"`                             // PICSAppUFS
	UpdatedAt                     time.Time `gorm:"updated_at"`                      //
	Version                       string    `gorm:"version"`                         //
	WishlistAvgPosition           float64   `gorm:"wishlist_avg_position"`           //
	WishlistCount                 int       `gorm:"wishlist_count"`                  //
}

func (app App) BSON() bson.D {

	return bson.D{
		{"achievements", app.Achievements},
		{"achievements_average_completion", app.AchievementsAverageCompletion},
		{"achievements_count", app.AchievementsCount},
		{"albummetadata", app.AlbumMetaData},
		{"background", app.Background},
		{"bundle_ids", app.BundleIDs},
		{"categories", app.Categories},
		{"change_number", app.ChangeNumber},
		{"change_number_date", app.ChangeNumberDate},
		{"client_icon", app.ClientIcon},
		{"coming_soon", app.ComingSoon},
		{"common", app.Common},
		{"config", app.Config},
		{"created_at", app.CreatedAt},
		{"demo_ids", app.DemoIDs},
		{"depots", app.Depots},
		{"developers", app.Developers},
		{"dlc", app.DLC},
		{"dlc_count", app.DLCCount},
		{"extended", app.Extended},
		{"game_id", app.GameID},
		{"game_name", app.GameName},
		{"genres", app.Genres},
		{"group_id", app.GroupID},
		{"group_followers", app.GroupFollowers},
		{"homepage", app.Homepage},
		{"icon", app.Icon},
		{"id", app.ID},
		{"install", app.Install},
		{"is_free", app.IsFree},
		{"items", app.Items},
		{"items_digest", app.ItemsDigest},
		{"launch", app.Launch},
		{"localization", app.Localization},
		{"logo", app.Logo},
		{"metacritic_score", app.MetacriticScore},
		{"metacritic_url", app.MetacriticURL},
		{"movies", app.Movies},
		{"name", app.Name},
		{"news_ids", app.NewsIDs},
		{"packages", app.Packages},
		{"platforms", app.Platforms},
		{"player_avg_week", app.PlayerAverageWeek},
		{"player_peak_alltime", app.PlayerPeakAllTime},
		{"player_peak_week", app.PlayerPeakWeek},
		{"player_trend", app.PlayerTrend},
		{"playtime_average", app.PlaytimeAverage},
		{"playtime_total", app.PlaytimeTotal},
		{"prices", app.Prices},
		{"public_only", app.PublicOnly},
		{"publishers", app.Publishers},
		{"related_app_ids", app.RelatedAppIDs},
		{"release_date", app.ReleaseDate},
		{"release_date_unix", app.ReleaseDateUnix},
		{"release_state", app.ReleaseState},
		{"reviews", app.Reviews},
		{"reviews_score", app.ReviewsScore},
		{"screenshots", app.Screenshots},
		{"description_short", app.ShortDescription},
		{"stats", app.Stats},
		{"steam_spy", app.SteamSpy},
		{"system_requirements", app.SystemRequirements},
		{"tags", app.Tags},
		{"twitch_id", app.TwitchID},
		{"twitch_url", app.TwitchURL},
		{"type", app.Type},
		{"ufs", app.UFS},
		{"updated_at", app.UpdatedAt},
		{"version", app.Version},
		{"wishlist_avg_position", app.WishlistAvgPosition},
		{"wishlist_count", app.WishlistCount},
	}
}

func (app App) Save() (err error) {

	if app.ID == 0 {
		return errors.New("invalid app id")
	}

	_, err = ReplaceOne(CollectionApps, bson.D{{"_id", app.ID}}, app)
	return err
}

func GetApp(id int) (app App, err error) {

	if !helpers.IsValidAppID(id) {
		return app, ErrInvalidAppID
	}

	err = FindOne(CollectionApps, bson.D{{"_id", id}}, nil, nil, &app)
	if err != nil {
		return app, err
	}
	if app.ID == 0 {
		return app, ErrNoDocuments
	}

	return app, err
}
