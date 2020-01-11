package mongo

import (
	"errors"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"go.mongodb.org/mongo-driver/bson"
)

var ErrInvalidAppID = errors.New("invalid app id")

type App struct {
	Achievements                  string             `bson:"achievements"`                    // []AppAchievement
	AchievementsAverageCompletion float64            `bson:"achievements_average_completion"` //
	AchievementsCount             int                `bson:"achievements_count"`              //
	AlbumMetaData                 string             `bson:"albummetadata"`                   // AlbumMetaData
	Background                    string             `bson:"background"`                      //
	BundleIDs                     []int              `bson:"bundle_ids"`                      //
	Categories                    []int              `bson:"categories"`                      //
	ChangeNumber                  int                `bson:"change_number"`                   //
	ChangeNumberDate              time.Time          `bson:"change_number_date"`              //
	ClientIcon                    string             `bson:"client_icon"`                     //
	ComingSoon                    bool               `bson:"coming_soon"`                     //
	Common                        map[string]string  `bson:"common"`                          //
	Config                        map[string]string  `bson:"config"`                          //
	CreatedAt                     time.Time          `bson:"created_at"`                      //
	DemoIDs                       []int              `bson:"demo_ids"`                        //
	Depots                        string             `bson:"depots"`                          // Depots
	Developers                    []int              `bson:"developers"`                      //
	DLC                           []int              `bson:"dlc"`                             //
	DLCCount                      int                `bson:"dlc_count"`                       //
	Extended                      map[string]string  `bson:"extended"`                        //
	GameID                        int                `bson:"game_id"`                         //
	GameName                      string             `bson:"game_name"`                       //
	Genres                        []int              `bson:"genres"`                          //
	GroupID                       string             `bson:"group_id"`                        //
	GroupFollowers                int                `bson:"group_followers"`                 //
	Homepage                      string             `bson:"homepage"`                        //
	Icon                          string             `bson:"icon"`                            //
	ID                            int                `bson:"id;primary_key"`                  //
	Install                       string             `bson:"install"`                         // map[string]interface{}
	IsFree                        bool               `bson:"is_free"`                         //
	Items                         int                `bson:"items"`                           //
	ItemsDigest                   string             `bson:"items_digest"`                    //
	Launch                        string             `bson:"launch"`                          // []db.PICSAppConfigLaunchItem
	Localization                  string             `bson:"localization"`                    // pics.Localisation
	Logo                          string             `bson:"logo"`                            //
	MetacriticScore               int8               `bson:"metacritic_score"`                //
	MetacriticURL                 string             `bson:"metacritic_url"`                  //
	Movies                        string             `bson:"movies"`                          // []AppVideo
	Name                          string             `bson:"name"`                            //
	NewsIDs                       []int64            `bson:"news_ids"`                        //
	Packages                      []int              `bson:"packages"`                        //
	Platforms                     []string           `bson:"platforms"`                       //
	PlayerAverageWeek             float64            `bson:"player_avg_week"`                 //
	PlayerPeakAllTime             int                `bson:"player_peak_alltime"`             //
	PlayerPeakWeek                int                `bson:"player_peak_week"`                //
	PlayerTrend                   int64              `bson:"player_trend"`                    //
	PlaytimeAverage               float64            `bson:"playtime_average"`                // Minutes
	PlaytimeTotal                 int64              `bson:"playtime_total"`                  // Minutes
	Prices                        string             `bson:"prices"`                          // ProductPrices
	PublicOnly                    bool               `bson:"public_only"`                     //
	Publishers                    []int              `bson:"publishers"`                      //
	RelatedAppIDs                 []int              `bson:"related_app_ids"`                 //
	RelatedOwnersAppIDs           []helpers.TupleInt `bson:"related_owners_app_ids"`          //
	ReleaseDate                   string             `bson:"release_date"`                    //
	ReleaseDateUnix               int64              `bson:"release_date_unix"`               //
	ReleaseState                  string             `bson:"release_state"`                   //
	Reviews                       string             `bson:"reviews"`                         // AppReviewSummary
	ReviewsScore                  float64            `bson:"reviews_score"`                   //
	Screenshots                   string             `bson:"screenshots"`                     // []AppImage
	ShortDescription              string             `bson:"description_short"`               //
	Stats                         string             `bson:"stats"`                           // []AppStat
	SteamSpy                      string             `bson:"steam_spy"`                       // AppSteamSpy
	SystemRequirements            string             `bson:"system_requirements"`             // map[string]interface{}
	Tags                          []int              `bson:"tags"`                            //
	TwitchID                      int                `bson:"twitch_id"`                       //
	TwitchURL                     string             `bson:"twitch_url"`                      //
	Type                          string             `bson:"type"`                            //
	UFS                           map[string]string  `bson:"ufs"`                             //
	UpdatedAt                     time.Time          `bson:"updated_at"`                      //
	Version                       string             `bson:"version"`                         //
	WishlistAvgPosition           float64            `bson:"wishlist_avg_position"`           //
	WishlistCount                 int                `bson:"wishlist_count"`                  //
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
		{"_id", app.ID},
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
