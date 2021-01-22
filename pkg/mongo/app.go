package mongo

import (
	"errors"
	"html/template"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mysql/pics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	PlatformWindows = "windows"
	PlatformMac     = "macos"
	PlatformLinux   = "linux"
)

var ErrInvalidAppID = errors.New("invalid app id")

type App struct {
	Achievements                  []helpers.Tuple                `bson:"achievements_5"` // Icon -> Title
	AchievementsAverageCompletion float64                        `bson:"achievements_average_completion"`
	AchievementsCount             int                            `bson:"achievements_count"`
	AchievementsCountTotal        int                            `bson:"achievements_count_total"` // Including inactive ones
	AlbumMetaData                 pics.AlbumMetaData             `bson:"albummetadata"`
	Background                    string                         `bson:"background"`
	BadgeOwners                   int64                          `bson:"badge_owners"`
	Bundles                       []int                          `bson:"bundle_ids"`
	Categories                    []int                          `bson:"categories"`
	ChangeNumber                  int                            `bson:"change_number"`
	ChangeNumberDate              time.Time                      `bson:"change_number_date"`
	ComingSoon                    bool                           `bson:"coming_soon"`
	Common                        pics.PICSKeyValues             `bson:"common"`
	Config                        pics.PICSKeyValues             `bson:"config"`
	CreatedAt                     time.Time                      `bson:"created_at"`
	Countries                     map[string]int                 `bson:"countries"`
	Demos                         []int                          `bson:"demo_ids"`
	Depots                        pics.Depots                    `bson:"depots"`
	Developers                    []int                          `bson:"developers"`
	DLCCount                      int                            `bson:"dlc_count"`
	Extended                      pics.PICSKeyValues             `bson:"extended"`
	GameID                        int                            `bson:"game_id"`
	GameName                      string                         `bson:"game_name"`
	Genres                        []int                          `bson:"genres"`
	GroupID                       string                         `bson:"group_id"`
	GroupFollowers                int                            `bson:"group_followers"`
	Homepage                      string                         `bson:"homepage"`
	Icon                          string                         `bson:"icon"`
	ID                            int                            `bson:"_id" json:"id"`
	Install                       map[string]interface{}         `bson:"install"`
	IsFree                        bool                           `bson:"is_free"`
	Items                         int                            `bson:"items"`
	ItemsDigest                   string                         `bson:"items_digest"`
	Launch                        []pics.PICSAppConfigLaunchItem `bson:"launch"`
	Localization                  pics.Localisation              `bson:"localization"`
	LocalizationCount             int                            `bson:"localization_count"`
	MetacriticScore               int8                           `bson:"metacritic_score"`
	MetacriticURL                 string                         `bson:"metacritic_url"`
	Movies                        []helpers.AppVideo             `bson:"movies"`
	Name                          string                         `bson:"name"`
	NewsIDs                       []int64                        `bson:"news_ids"`
	Owners                        int64                          `bson:"owners"` // On Global Steam
	Packages                      []int                          `bson:"packages"`
	Platforms                     []string                       `bson:"platforms"`
	PlayerAverageWeek             float64                        `bson:"player_avg_week"`
	PlayerPeakAllTime             int                            `bson:"player_peak_alltime"`
	PlayerPeakAllTimeTime         time.Time                      `bson:"player_peak_alltime_time"`
	PlayerPeakWeek                int                            `bson:"player_peak_week"`
	PlayerTrend                   float64                        `bson:"player_trend"`
	PlaytimeAverage               float64                        `bson:"playtime_average"` // Minutes
	PlaytimeTotal                 int64                          `bson:"playtime_total"`   // Minutes
	Prices                        helpers.ProductPrices          `bson:"prices"`
	PublicOnly                    bool                           `bson:"public_only"`
	Publishers                    []int                          `bson:"publishers"`
	RelatedAppIDs                 []int                          `bson:"related_app_ids"`             // Taken from store page
	RelatedOwnersAppIDsDate       time.Time                      `bson:"related_owners_app_ids_date"` // Calculated from owners - Last Updated
	ReleaseDate                   string                         `bson:"release_date"`                // Steam release
	ReleaseDateUnix               int64                          `bson:"release_date_unix"`           // Steam release
	ReleaseDateOriginal           int64                          `bson:"release_date_original"`       // Game release
	ReleaseState                  string                         `bson:"release_state"`
	Reviews                       helpers.AppReviewSummary       `bson:"reviews"`
	ReviewsScore                  float64                        `bson:"reviews_score"`
	ReviewsCount                  int                            `bson:"reviews_count"`
	Screenshots                   []helpers.AppImage             `bson:"screenshots"`
	ShortDescription              string                         `bson:"description_short"`
	Stats                         []helpers.AppStat              `bson:"stats"`
	SteamSpy                      helpers.AppSteamSpy            `bson:"steam_spy"`
	SystemRequirements            map[string]interface{}         `bson:"system_requirements"`
	Tags                          []int                          `bson:"tags"`       // Current tags, used for filtering
	TagCounts                     []AppTagCount                  `bson:"tag_counts"` // Includes old tags
	TwitchID                      int                            `bson:"twitch_id"`
	TwitchURL                     string                         `bson:"twitch_url"`
	Type                          string                         `bson:"type"`
	UFS                           pics.PICSKeyValues             `bson:"ufs"`
	UpdatedAt                     time.Time                      `bson:"updated_at"`
	Version                       string                         `bson:"version"`
	WishlistAvgPosition           float64                        `bson:"wishlist_avg_position"`
	WishlistCount                 int                            `bson:"wishlist_count"`
	WishlistPercent               float64                        `bson:"wishlist_percent"`
	WishlistFirsts                float64                        `bson:"wishlist_firsts"`
}

func (app App) BSON() bson.D {

	app.UpdatedAt = time.Now()
	app.ReleaseState = strings.ToLower(app.ReleaseState)

	sort.Slice(app.TagCounts, func(i, j int) bool {
		return app.TagCounts[i].Count > app.TagCounts[j].Count
	})

	return bson.D{
		{"achievements_5", app.Achievements},
		{"achievements_average_completion", app.AchievementsAverageCompletion},
		{"achievements_count", app.AchievementsCount},
		{"achievements_count_total", app.AchievementsCountTotal},
		{"albummetadata", app.AlbumMetaData},
		{"background", app.Background},
		{"badge_owners", app.BadgeOwners},
		{"bundle_ids", app.Bundles},
		{"categories", app.Categories},
		{"change_number", app.ChangeNumber},
		{"change_number_date", app.ChangeNumberDate},
		{"coming_soon", app.ComingSoon},
		{"common", app.Common},
		{"config", app.Config},
		{"created_at", app.CreatedAt},
		{"countries", app.Countries},
		{"demo_ids", app.Demos},
		{"depots", app.Depots},
		{"developers", app.Developers},
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
		{"localization_count", app.LocalizationCount},
		{"metacritic_score", app.MetacriticScore},
		{"metacritic_url", app.MetacriticURL},
		{"movies", app.Movies},
		{"name", app.Name},
		{"news_ids", app.NewsIDs},
		{"owners", app.Owners},
		{"packages", app.Packages},
		{"platforms", app.Platforms},
		{"player_avg_week", app.PlayerAverageWeek},
		{"player_peak_alltime", app.PlayerPeakAllTime},
		{"player_peak_alltime_time", app.PlayerPeakAllTimeTime},
		{"player_peak_week", app.PlayerPeakWeek},
		{"player_trend", app.PlayerTrend},
		{"playtime_average", app.PlaytimeAverage},
		{"playtime_total", app.PlaytimeTotal},
		{"prices", app.Prices},
		{"public_only", app.PublicOnly},
		{"publishers", app.Publishers},
		{"related_app_ids", app.RelatedAppIDs},
		{"related_owners_app_ids_date", app.RelatedOwnersAppIDsDate},
		{"release_date", app.ReleaseDate},
		{"release_date_unix", app.ReleaseDateUnix},
		{"release_date_original", app.ReleaseDateOriginal},
		{"release_state", app.ReleaseState},
		{"reviews", app.Reviews},
		{"reviews_score", app.ReviewsScore},
		{"reviews_count", app.ReviewsCount},
		{"screenshots", app.Screenshots},
		{"description_short", app.ShortDescription},
		{"stats", app.Stats},
		{"steam_spy", app.SteamSpy},
		{"system_requirements", app.SystemRequirements},
		{"tags", app.Tags},
		{"tag_counts", app.TagCounts},
		{"twitch_id", app.TwitchID},
		{"twitch_url", app.TwitchURL},
		{"type", app.Type},
		{"ufs", app.UFS},
		{"updated_at", app.UpdatedAt},
		{"version", app.Version},
		{"wishlist_avg_position", app.WishlistAvgPosition},
		{"wishlist_count", app.WishlistCount},
		{"wishlist_percent", app.WishlistPercent},
		{"wishlist_firsts", app.WishlistFirsts},
	}
}

func (app App) GetID() int {
	return app.ID
}

func (app App) GetPlayersPeakWeek() int {
	return app.PlayerPeakWeek
}

func (app App) GetProductType() helpers.ProductType {
	return helpers.ProductTypeApp
}

func (app App) GetPrices() (prices helpers.ProductPrices) {
	return app.Prices
}

func (app App) GetName() string {
	return helpers.GetAppName(app.ID, app.Name)
}

func (app App) GetIcon() string {
	return helpers.GetAppIcon(app.ID, app.Icon)
}

func (app App) GetPath() string {
	return helpers.GetAppPath(app.ID, app.Name)
}

func (app App) GetTrend() string {
	return helpers.GetTrendValue(app.PlayerTrend)
}

func (app App) GetType() string {
	return helpers.GetAppType(app.Type)
}

// For an interface
func (app App) GetBackground() string {
	return app.Background
}

func (app App) GetTypeLower() (ret string) {

	switch app.Type {
	case "dlc":
		return "DLC"
	case "advertising":
		return "advertisement"
	case "music":
		return "soundstracks"
	case "":
		return "app"
	default:
		return strings.ToLower(app.Type)
	}
}

func (app App) GetStoreLink() string {
	return helpers.GetAppStoreLink(app.ID)
}

// 460 x 215
func (app App) GetHeaderImage() string {
	return "https://steamcdn-a.akamaihd.net/steam/apps/" + strconv.Itoa(app.ID) + "/header.jpg"
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
	for _, tagID := range app.Tags {
		if val, ok := tagMap[tagID]; ok {
			coopTags = append(coopTags, val)
		}
	}

	return strings.Join(coopTags, ", "), nil
}

// func (app App) GetHeaderImage2() string {
//
// 	params := url.Values{}
// 	params.Set("url", app.GetHeaderImage())
// 	params.Set("q", "10")
// 	params.Set("output", "webp")
//
// 	return "https://images.weserv.nl?" + params.Encode()
// }

func (app App) GetCommunityLink() string {
	return helpers.GetAppCommunityLink(app.ID)
}

func (app App) GetInstallLink() template.URL {
	return template.URL("steam://install/" + strconv.Itoa(app.ID))
}

func (app App) GetPlayLink() template.URL {
	return helpers.GetAppPlayLink(app.ID)
}

func (app App) GetReviewScore() string {
	return helpers.GetAppReviewScore(app.ReviewsScore)
}

func (app App) GetFollowers() string {
	return helpers.GetAppFollowers(app.GroupID, app.GroupFollowers)
}

func (app App) GetPlatformImages() (ret template.HTML) {

	if len(app.Platforms) == 0 {
		return ""
	}

	if helpers.SliceHasString(PlatformWindows, app.Platforms) {
		ret = ret + `<a href="/games?platforms=windows"><i class="fab fa-windows" data-toggle="tooltip" data-placement="top" title="Windows"></i></a>`
	} else {
		ret = ret + `<span class="space"></span>`
	}

	if helpers.SliceHasString(PlatformMac, app.Platforms) {
		ret = ret + `<a href="/games?platforms=macos"><i class="fab fa-apple" data-toggle="tooltip" data-placement="top" title="Mac"></i></a>`
	} else {
		ret = ret + `<span class="space"></span>`
	}

	if helpers.SliceHasString(PlatformLinux, app.Platforms) {
		ret = ret + `<a href="/games?platforms=linux"><i class="fab fa-linux" data-toggle="tooltip" data-placement="top" title="Linux"></i></a>`
	} else {
		ret = ret + `<span class="space"></span>`
	}

	return ret
}

func (app App) GetMetaImage() string {

	if len(app.Screenshots) == 0 {
		return app.GetHeaderImage()
	}
	return app.Screenshots[0].PathFull
}

func (app App) GetMicroTrailer() string {

	if len(app.Movies) == 0 {
		return ""
	}
	return app.Movies[0].Micro()
}

func (app App) GetPlaytimeAverage() string {

	x := int(math.Round(app.PlaytimeAverage))
	return helpers.GetTimeLong(x, 3)
}

func (app App) GetAppRelatedApps() (apps []App, err error) {

	apps = []App{} // Needed for marshalling into type

	if len(app.RelatedAppIDs) == 0 {
		return apps, nil
	}

	err = memcache.GetSetInterface(memcache.ItemAppRelated(app.ID), &apps, func() (interface{}, error) {

		return GetAppsByID(app.RelatedAppIDs, bson.M{"_id": 1, "name": 1, "icon": 1, "tags": 1})
	})

	return apps, err
}

func (app App) GetDemos() (demos []App, err error) {

	demos = []App{} // Needed for marshalling into type

	if len(app.Demos) == 0 {
		return demos, nil
	}

	err = memcache.GetSetInterface(memcache.ItemAppDemos(app.ID), &demos, func() (interface{}, error) {
		return GetAppsByID(app.Demos, bson.M{"_id": 1, "name": 1})
	})

	return demos, err
}

func (app App) ReadPICS(m map[string]string) (config pics.PICSKeyValues) {

	config = pics.PICSKeyValues{}

	for k, v := range m {
		config[k] = v
	}

	return config
}

func (app App) GetReleaseDateNice() string {
	return helpers.GetAppReleaseDateNice(app.ReleaseDateOriginal, app.ReleaseDateUnix, app.ReleaseDate)
}

func (app App) GetSteamPricesURL() string {

	switch app.Type {
	case "game":
		return "app"
	case "dlc":
		return "dlc"
	case "application":
		return "sw"
	case "hardware":
		return "hw"
	default:
		return ""
	}
}

var GetPlayersInGameLock sync.Mutex

func (app App) GetPlayersInGame() (players int64, err error) {

	GetPlayersInGameLock.Lock()
	defer GetPlayersInGameLock.Unlock()

	err = memcache.GetSetInterface(memcache.ItemAppPlayersRow(app.ID), &players, func() (interface{}, error) {

		builder := influxql.NewBuilder()
		builder.AddSelect("player_count", "")
		builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
		builder.AddWhere("app_id", "=", app.ID)
		builder.AddOrderBy("time", false)
		builder.SetLimit(1)

		return influx.GetFirstInfluxInt(builder)
	})

	return players, err
}

// Only used for steam app 0
var GetPlayersOnlineLock sync.Mutex

func (app App) GetPlayersOnline() (players int64, err error) {

	GetPlayersOnlineLock.Lock()
	defer GetPlayersOnlineLock.Unlock()

	err = memcache.GetSetInterface(memcache.ItemAppPlayersInGameRow, &players, func() (interface{}, error) {

		builder := influxql.NewBuilder()
		builder.AddSelect("player_online", "")
		builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
		builder.AddWhere("app_id", "=", 0)
		builder.AddOrderBy("time", false)
		builder.SetLimit(1)

		return influx.GetFirstInfluxInt(builder)
	})

	return players, err
}

func (app App) GetSystemRequirements() (ret []helpers.SystemRequirement) {

	flattened := helpers.FlattenMap(app.SystemRequirements)

	for k, v := range flattened {
		if val, ok := v.(string); ok {
			ret = append(ret, helpers.SystemRequirement{Key: k, Val: val})
		}
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Key < ret[j].Key
	})

	return ret
}

func (app App) GetPICSUpdatedNice() string {

	if app.ChangeNumberDate.IsZero() || app.ChangeNumberDate.Unix() == 0 {
		return "-"
	}

	return app.ChangeNumberDate.Format(helpers.DateYearTime)
}

func (app App) GetUpdatedNice() string {
	return app.UpdatedAt.Format(helpers.DateYearTime)
}

func (app App) GetPeakTimeNice() string {
	if app.PlayerPeakAllTimeTime.IsZero() {
		return "-"
	}
	return app.PlayerPeakAllTimeTime.Format(helpers.DateYearTime)
}

func (app App) ShouldUpdate() bool {

	return app.UpdatedAt.Before(time.Now().Add(time.Hour * 24 * -1))
}

func (app App) GetTags() (stats []Stat, err error) {
	return GetStatsByType(StatsTypeTags, app.Tags, app.ID)
}

func (app App) GetCategories() (stats []Stat, err error) {
	return GetStatsByType(StatsTypeCategories, app.Categories, app.ID)
}

func (app App) GetGenres() (stats []Stat, err error) {
	return GetStatsByType(StatsTypeGenres, app.Genres, app.ID)
}

func (app App) GetPublishers() (stats []Stat, err error) {
	return GetStatsByType(StatsTypePublishers, app.Publishers, app.ID)
}

func (app App) GetDevelopers() (stats []Stat, err error) {
	return GetStatsByType(StatsTypeDevelopers, app.Developers, app.ID)
}

//
type AppTagCount struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func (t AppTagCount) GetPath() string {
	return helpers.GetStatPath(StatsTypeTags.MongoCol(), t.ID, t.Name)
}

//
func ChunkApps(strings []App, n int) (chunks [][]App) {

	for i := 0; i < len(strings); i += n {
		end := i + n

		if end > len(strings) {
			end = len(strings)
		}

		chunks = append(chunks, strings[i:end])
	}
	return chunks
}

func UpdateAppsInflux(writes []mongo.WriteModel) (err error) {

	if len(writes) == 0 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	c := client.Database(config.C.MongoDatabase).Collection(CollectionApps.String())
	_, err = c.BulkWrite(ctx, writes, options.BulkWrite())

	return err
}

func ensureAppIndexes() {

	var indexModels = []mongo.IndexModel{
		{Keys: bson.D{{"achievements_average_completion", -1}}},
		{Keys: bson.D{{"achievements_count", -1}, {"achievements_average_completion", -1}}},
		{Keys: bson.D{{"categories", 1}}},
		{Keys: bson.D{{"developers", 1}}},
		{Keys: bson.D{{"genres", 1}}},
		{Keys: bson.D{{"genres", -1}}},
		{Keys: bson.D{{"group_followers", -1}}},
		{Keys: bson.D{{"player_peak_week", -1}}},
		{Keys: bson.D{{"player_trend", -1}}},
		{Keys: bson.D{{"publishers", 1}}},
		{Keys: bson.D{{"release_date_unix", 1}}},
		{Keys: bson.D{{"release_date_unix", -1}}},
		{Keys: bson.D{{"reviews_score", -1}}},
		{Keys: bson.D{{"tags", 1}}},
		{Keys: bson.D{{"tags", -1}}},
		{Keys: bson.D{{"type", 1}}},
		{Keys: bson.D{{"wishlist_avg_position", 1}}},
		{Keys: bson.D{{"wishlist_count", -1}}},
	}

	//
	client, ctx, err := getMongo()
	if err != nil {
		log.ErrS(err)
		return
	}

	_, err = client.Database(config.C.MongoDatabase).Collection(CollectionApps.String()).Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		log.ErrS(err)
	}
}

func BatchApps(filter bson.D, projection bson.M, callback func(apps []App)) (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		apps, err := GetApps(offset, limit, bson.D{{"_id", 1}}, filter, projection)
		if err != nil {
			return err
		}

		callback(apps)

		if int64(len(apps)) != limit {
			break
		}

		offset += limit
	}

	return nil
}

func GetApp(id int, full ...bool) (app App, err error) {

	if !helpers.IsValidAppID(id) {
		return app, ErrInvalidAppID
	}

	if len(full) > 0 && full[0] {

		// Load from Mongo
		err = FindOne(CollectionApps, bson.D{{"_id", id}}, nil, nil, &app)

	} else {

		// Load from Memcache
		err = memcache.GetSetInterface(memcache.ItemApp(id), &app, func() (interface{}, error) {

			var projection = bson.M{"reviews.reviews": 0, "localization": 0, "achievements": 0, "achievements_5": 0} // Too much for memcache

			err = FindOne(CollectionApps, bson.D{{"_id", id}}, nil, projection, &app)
			return app, err
		})
	}

	return app, err
}

func GetApps(offset int64, limit int64, sort bson.D, filter bson.D, projection bson.M) (apps []App, err error) {

	cur, ctx, err := find(CollectionApps, offset, limit, sort, filter, projection, nil)
	if err != nil {
		return apps, err
	}

	defer closeCursor(cur, ctx)

	for cur.Next(ctx) {

		var app App
		err := cur.Decode(&app)
		if err != nil {
			log.ErrS(err, app.ID)
		} else {
			apps = append(apps, app)
		}
	}

	return apps, cur.Err()
}

func GetAppsByID(ids []int, projection bson.M) (apps []App, err error) {

	if len(ids) < 1 {
		return apps, nil
	}

	a := bson.A{}
	for _, v := range ids {
		a = append(a, v)
	}

	return GetApps(0, 0, nil, bson.D{{"_id", bson.M{"$in": a}}}, projection)
}

func PopularApps() (apps []App, err error) {

	err = memcache.GetSetInterface(memcache.ItemAppsPopular, &apps, func() (interface{}, error) {

		return GetApps(
			0,
			50,
			bson.D{{"player_peak_week", -1}},
			bson.D{{"type", "game"}},
			bson.M{"_id": 1, "name": 1, "player_peak_week": 1, "background": 1},
		)
	})

	return apps, err
}

func PopularNewApps() (apps []App, err error) {

	err = memcache.GetSetInterface(memcache.ItemNewPopularApps, &apps, func() (interface{}, error) {

		releaseDate := time.Now().AddDate(0, 0, -config.C.NewReleaseDays).Unix()

		return GetApps(
			0,
			10, // Keep even
			bson.D{{Key: "player_peak_week", Value: -1}},
			bson.D{
				{Key: "release_date_unix", Value: bson.M{"$gt": releaseDate}},
				{Key: "type", Value: "game"},
			},
			bson.M{"_id": 1, "name": 1, "player_peak_week": 1},
		)
	})

	return apps, err
}

func TrendingApps() (apps []App, err error) {

	err = memcache.GetSetInterface(memcache.ItemAppsTrending, &apps, func() (interface{}, error) {
		return GetApps(
			0,
			10,
			bson.D{{"player_trend", -1}},
			nil,
			bson.M{"_id": 1, "name": 1, "player_trend": 1},
		)
	})

	return apps, err
}

func GetAppsGroupedByType(code steamapi.ProductCC) (counts []AppTypeCount, err error) {

	err = memcache.GetSetInterface(memcache.ItemAppTypeCounts(code), &counts, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return counts, err
		}

		pipeline := mongo.Pipeline{
			{{Key: "$sort", Value: bson.M{"type": 1}}}, // Just here to hit the index
			{{Key: "$group", Value: bson.M{
				"_id":   "$type",
				"count": bson.M{"$sum": 1},
				"value": bson.M{"$sum": "$prices." + code + ".final"},
			}}},
		}

		cur, err := client.Database(config.C.MongoDatabase, options.Database()).Collection(CollectionApps.String()).Aggregate(ctx, pipeline, options.Aggregate())
		if err != nil {
			return counts, err
		}

		defer closeCursor(cur, ctx)

		var unknown int64
		var counts []AppTypeCount
		for cur.Next(ctx) {

			var appType AppTypeCount
			err := cur.Decode(&appType)
			if err != nil {
				log.ErrS(err, appType.Type)
			}

			if appType.Type == "" {
				unknown = appType.Count
			} else {
				counts = append(counts, appType)
			}
		}

		sort.Slice(counts, func(i, j int) bool {
			return counts[i].Count > counts[j].Count
		})

		counts = append(counts, AppTypeCount{Count: unknown})

		return counts, cur.Err()
	})

	return counts, err
}

type AppTypeCount struct {
	Type  string `json:"type" bson:"_id"`
	Count int64  `json:"count" bson:"count"`
	Value int64  `json:"value" bson:"value"`
}

// Used in templates
func (atc AppTypeCount) Format() string {
	return helpers.GetAppType(atc.Type)
}

func GetAppsGroupedByReleaseDate() (counts []AppReleaseDateCount, err error) {

	err = memcache.GetSetInterface(memcache.ItemAppReleaseDateCounts, &counts, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return counts, err
		}

		pipeline := mongo.Pipeline{
			bson.D{{Key: "$match", Value: bson.M{"release_date_unix": bson.M{"$gte": time.Now().AddDate(-2, 0, 0).Unix()}}}},
			bson.D{{Key: "$match", Value: bson.M{"release_date_unix": bson.M{"$lte": time.Now().AddDate(0, 0, 1).Unix()}}}},
			bson.D{{Key: "$project", Value: bson.M{"release_date": bson.M{"$toDate": bson.M{"$multiply": bson.A{"$release_date_unix", 1000}}}}}},
			bson.D{{Key: "$group", Value: bson.M{"_id": bson.M{"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$release_date"}}, "count": bson.M{"$sum": 1}}}},
		}

		cur, err := client.Database(config.C.MongoDatabase, options.Database()).Collection(CollectionApps.String()).Aggregate(ctx, pipeline, options.Aggregate())
		if err != nil {
			return counts, err
		}

		defer closeCursor(cur, ctx)

		var counts []AppReleaseDateCount
		for cur.Next(ctx) {

			var row AppReleaseDateCount
			err := cur.Decode(&row)
			if err != nil {
				log.ErrS(err, row.Date)
			}

			counts = append(counts, row)
		}

		sort.Slice(counts, func(i, j int) bool {
			return counts[i].Date < counts[j].Date
		})

		return counts, cur.Err()
	})

	return counts, err
}

type AppReleaseDateCount struct {
	Date  string `json:"date" bson:"_id"`
	Count int64  `json:"count" bson:"count"`
}

func GetAppsGroupedByReviewScore() (counts []AppReviewScoreCount, err error) {

	err = memcache.GetSetInterface(memcache.ItemAppReviewScoreCounts, &counts, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return counts, err
		}

		pipeline := mongo.Pipeline{
			bson.D{{Key: "$match", Value: bson.M{"reviews_score": bson.M{"$gt": 0}}}},
			bson.D{{Key: "$group", Value: bson.M{"_id": bson.M{"$floor": "$reviews_score"}, "count": bson.M{"$sum": 1}}}},
		}

		cur, err := client.Database(config.C.MongoDatabase, options.Database()).Collection(CollectionApps.String()).Aggregate(ctx, pipeline, options.Aggregate())
		if err != nil {
			return counts, err
		}

		defer closeCursor(cur, ctx)

		var counts []AppReviewScoreCount
		for cur.Next(ctx) {

			var row AppReviewScoreCount
			err := cur.Decode(&row)
			if err != nil {
				log.ErrS(err, row.Score)
			}

			counts = append(counts, row)
		}

		sort.Slice(counts, func(i, j int) bool {
			return counts[i].Score < counts[j].Score
		})

		return counts, cur.Err()
	})

	return counts, err
}

type AppReviewScoreCount struct {
	Score int64 `json:"score" bson:"_id"`
	Count int64 `json:"count" bson:"count"`
}
