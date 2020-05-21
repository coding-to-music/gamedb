package mongo

import (
	"errors"
	"html/template"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/i18n"
	"github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql/pics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	platformWindows = "windows"
	platformMac     = "macos"
	platformLinux   = "linux"
)

var ErrInvalidAppID = errors.New("invalid app id")

type App struct {
	Achievements                  []AppAchievement               `bson:"achievements_5"` // The first 5 only
	AchievementsAverageCompletion float64                        `bson:"achievements_average_completion"`
	AchievementsCount             int                            `bson:"achievements_count"`
	AchievementsCountTotal        int                            `bson:"achievements_count_total"` // Including inactive ones
	AlbumMetaData                 pics.AlbumMetaData             `bson:"albummetadata"`
	Background                    string                         `bson:"background"`
	Bundles                       []int                          `bson:"bundle_ids"`
	Categories                    []int                          `bson:"categories"`
	ChangeNumber                  int                            `bson:"change_number"`
	ChangeNumberDate              time.Time                      `bson:"change_number_date"`
	ClientIcon                    string                         `bson:"client_icon"`
	ComingSoon                    bool                           `bson:"coming_soon"`
	Common                        pics.PICSKeyValues             `bson:"common"`
	Config                        pics.PICSKeyValues             `bson:"config"`
	CreatedAt                     time.Time                      `bson:"created_at"`
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
	Logo                          string                         `bson:"logo"`
	MetacriticScore               int8                           `bson:"metacritic_score"`
	MetacriticURL                 string                         `bson:"metacritic_url"`
	Movies                        []helpers.AppVideo             `bson:"movies"`
	Name                          string                         `bson:"name"`
	NewsIDs                       []int64                        `bson:"news_ids"`
	Packages                      []int                          `bson:"packages"`
	Platforms                     []string                       `bson:"platforms"`
	PlayerAverageWeek             float64                        `bson:"player_avg_week"`
	PlayerPeakAllTime             int                            `bson:"player_peak_alltime"`
	PlayerPeakWeek                int                            `bson:"player_peak_week"`
	PlayerTrend                   int64                          `bson:"player_trend"`
	PlaytimeAverage               float64                        `bson:"playtime_average"` // Minutes
	PlaytimeTotal                 int64                          `bson:"playtime_total"`   // Minutes
	Prices                        helpers.ProductPrices          `bson:"prices"`
	PublicOnly                    bool                           `bson:"public_only"`
	Publishers                    []int                          `bson:"publishers"`
	RelatedAppIDs                 []int                          `bson:"related_app_ids"`        // Taken from store page
	RelatedOwnersAppIDs           []helpers.TupleInt             `bson:"related_owners_app_ids"` // Calculated from owners
	ReleaseDate                   string                         `bson:"release_date"`           // Steam release
	ReleaseDateUnix               int64                          `bson:"release_date_unix"`      // Steam release
	ReleaseDateOriginal           int64                          `bson:"release_date_original"`  // Game release
	ReleaseState                  string                         `bson:"release_state"`
	Reviews                       helpers.AppReviewSummary       `bson:"reviews"`
	ReviewsScore                  float64                        `bson:"reviews_score"`
	ReviewsCount                  int                            `bson:"reviews_count"`
	Screenshots                   []helpers.AppImage             `bson:"screenshots"`
	ShortDescription              string                         `bson:"description_short"`
	Stats                         []helpers.AppStat              `bson:"stats"`
	SteamSpy                      helpers.AppSteamSpy            `bson:"steam_spy"`
	SystemRequirements            map[string]interface{}         `bson:"system_requirements"`
	Tags                          []int                          `bson:"tags"`
	TagCounts                     []AppTagCount                  `bson:"tag_counts"`
	TwitchID                      int                            `bson:"twitch_id"`
	TwitchURL                     string                         `bson:"twitch_url"`
	Type                          string                         `bson:"type"`
	UFS                           pics.PICSKeyValues             `bson:"ufs"`
	UpdatedAt                     time.Time                      `bson:"updated_at"`
	Version                       string                         `bson:"version"`
	WishlistAvgPosition           float64                        `bson:"wishlist_avg_position"`
	WishlistCount                 int                            `bson:"wishlist_count"`
	Score                         float64                        `bson:"score,omitempty"` //  Just used for search
}

func (app App) BSON() bson.D {

	app.UpdatedAt = time.Now()
	app.ReleaseState = strings.ToLower(app.ReleaseState)

	return bson.D{
		{"achievements_5", app.Achievements},
		{"achievements_average_completion", app.AchievementsAverageCompletion},
		{"achievements_count", app.AchievementsCount},
		{"achievements_count_total", app.AchievementsCountTotal},
		{"albummetadata", app.AlbumMetaData},
		{"background", app.Background},
		{"bundle_ids", app.Bundles},
		{"categories", app.Categories},
		{"change_number", app.ChangeNumber},
		{"change_number_date", app.ChangeNumberDate},
		{"client_icon", app.ClientIcon},
		{"coming_soon", app.ComingSoon},
		{"common", app.Common},
		{"config", app.Config},
		{"created_at", app.CreatedAt},
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
		{"related_owners_app_ids", app.RelatedOwnersAppIDs},
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
	}
}

type AppTagCount struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func (app App) GetID() int {
	return app.ID
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

func (app App) GetIcon() (ret string) {
	return helpers.GetAppIcon(app.ID, app.Icon)
}

func (app App) GetPath() string {
	return helpers.GetAppPath(app.ID, app.Name)
}

func (app App) GetType() (ret string) {
	return helpers.GetAppType(app.Type)
}

func (app App) GetTypeLower() (ret string) {

	switch app.Type {
	case "dlc":
		return "DLC"
	case "":
		return "app"
	default:
		return strings.ToLower(app.Type)
	}
}

func (app App) GetStoreLink() string {
	return helpers.GetAppStoreLink(app.ID)
}

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
	name := config.Config.GameDBShortName.Get()
	return "https://steamcommunity.com/app/" + strconv.Itoa(app.ID) + "?utm_source=" + name + "&utm_medium=link&curator_clanid=" // todo curator_clanid
}

func (app App) GetInstallLink() template.URL {
	return template.URL("steam://install/" + strconv.Itoa(app.ID))
}

func (app App) GetPlayLink() template.URL {
	return template.URL("steam://run/" + strconv.Itoa(app.ID))
}

func (app App) GetReviewScore() string {
	return helpers.FloatToString(app.ReviewsScore, 2) + "%"
}

func (app App) GetFollowers() (ret string) {

	if app.GroupID == "" {
		return "-"
	}

	return humanize.Comma(int64(app.GroupFollowers))
}

func (app App) GetPlatformImages() (ret template.HTML, err error) {

	if len(app.Platforms) == 0 {
		return "", nil
	}

	if helpers.SliceHasString(platformWindows, app.Platforms) {
		ret = ret + `<a href="/apps?platforms=windows"><i class="fab fa-windows" data-toggle="tooltip" data-placement="top" title="Windows"></i></a>`
	} else {
		ret = ret + `<span class="space"></span>`
	}

	if helpers.SliceHasString(platformMac, app.Platforms) {
		ret = ret + `<a href="/apps?platforms=macos"><i class="fab fa-apple" data-toggle="tooltip" data-placement="top" title="Mac"></i></a>`
	} else {
		ret = ret + `<span class="space"></span>`
	}

	if helpers.SliceHasString(platformLinux, app.Platforms) {
		ret = ret + `<a href="/apps?platforms=linux"><i class="fab fa-linux" data-toggle="tooltip" data-placement="top" title="Linux"></i></a>`
	} else {
		ret = ret + `<span class="space"></span>`
	}

	return ret, nil
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

	var item = memcache.MemcacheAppRelated(app.ID)

	err = memcache.GetSetInterface(item.Key, item.Expiration, &apps, func() (interface{}, error) {

		return GetAppsByID(app.RelatedAppIDs, bson.M{"_id": 1, "name": 1, "icon": 1, "tags": 1})
	})

	return apps, err
}

func (app App) GetDemos() (demos []App, err error) {

	demos = []App{} // Needed for marshalling into type

	if len(app.Demos) == 0 {
		return demos, nil
	}

	var item = memcache.MemcacheAppDemos(app.ID)

	err = memcache.GetSetInterface(item.Key, item.Expiration, &demos, func() (interface{}, error) {
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

	var item = memcache.MemcacheAppPlayersRow(app.ID)

	err = memcache.GetSetInterface(item.Key, item.Expiration, &players, func() (interface{}, error) {

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

// Only used for steam app 0
var GetPlayersOnlineLock sync.Mutex

func (app App) GetPlayersOnline() (players int64, err error) {

	GetPlayersOnlineLock.Lock()
	defer GetPlayersOnlineLock.Unlock()

	var item = memcache.MemcacheAppPlayersInGameRow

	err = memcache.GetSetInterface(item.Key, item.Expiration, &players, func() (interface{}, error) {

		builder := influxql.NewBuilder()
		builder.AddSelect("player_online", "")
		builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
		builder.AddWhere("app_id", "=", 0)
		builder.AddOrderBy("time", false)
		builder.SetLimit(1)

		resp, err := influx.InfluxQuery(builder.String())

		return influx.GetFirstInfluxInt(resp), err
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

func (app App) ShouldUpdate() bool {

	return app.UpdatedAt.Before(time.Now().Add(time.Hour * 24 * -1))
}

func (app App) Save() (err error) {

	if !helpers.IsValidAppID(app.ID) {
		return errors.New("invalid app id")
	}

	_, err = ReplaceOne(CollectionApps, bson.D{{"_id", app.ID}}, app)
	return err
}

func (app App) GetAppPackages() (packages []Package, err error) {

	packages = []Package{} // Needed for marshalling into type

	if len(app.Packages) == 0 {
		return packages, nil
	}

	var item = memcache.MemcacheAppPackages(app.ID)

	err = memcache.GetSetInterface(item.Key, item.Expiration, &packages, func() (interface{}, error) {
		return GetPackagesByID(app.Packages, bson.M{})
	})

	return packages, err
}

func CreateAppIndexes() {

	var indexModels []mongo.IndexModel

	var cols = []string{
		"achievements_average_completion",
		"achievements_count",
		"categories",
		"developers",
		"genres",
		"group_followers",
		"platforms",
		"player_peak_week",
		"player_trend",
		"publishers",
		"release_date_unix",
		"reviews_score",
		"tags",
		"type",
		"wishlist_avg_position",
		"wishlist_count",
	}

	// Price fields
	for _, v := range i18n.GetProdCCs(true) {
		cols = append(cols, "prices."+string(v.ProductCode)+".final")
	}

	//
	for _, v := range cols {
		indexModels = append(indexModels, mongo.IndexModel{
			Keys: bson.D{{v, 1}},
		}, mongo.IndexModel{
			Keys: bson.D{{v, -1}},
		})
	}

	// Text index
	indexModels = append(indexModels, mongo.IndexModel{
		Keys:    bson.D{{"name", "text"}},
		Options: options.Index().SetName("text"),
	})

	// Search
	indexModels = append(indexModels, mongo.IndexModel{
		Keys: bson.D{{"extended.aliases", 1}},
	})

	// Achievements page
	indexModels = append(indexModels, mongo.IndexModel{
		Keys: bson.D{{"achievements_count", -1}, {"achievements_average_completion", -1}},
	})

	//
	client, ctx, err := getMongo()
	if err != nil {
		log.Err(err)
		return
	}

	_, err = client.Database(MongoDatabase).Collection(CollectionApps.String()).Indexes().CreateMany(ctx, indexModels)
	log.Err(err)
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
		var item = memcache.MemcacheApp(id)
		err = memcache.GetSetInterface(item.Key, item.Expiration, &app, func() (interface{}, error) {

			var projection = bson.M{"reviews.reviews": 0, "localization": 0, "achievements": 0, "achievements_5": 0} // Too much for memcache

			err = FindOne(CollectionApps, bson.D{{"_id", id}}, nil, projection, &app)
			return app, err
		})
	}

	return app, err
}

func GetApps(offset int64, limit int64, sort bson.D, filter bson.D, projection bson.M, ops *options.FindOptions) (apps []App, err error) {

	cur, ctx, err := Find(CollectionApps, offset, limit, sort, filter, projection, nil)
	if err != nil {
		return apps, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var app App
		err := cur.Decode(&app)
		if err != nil {
			log.Err(err, app.ID)
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

	return GetApps(0, 0, nil, bson.D{{"_id", bson.M{"$in": a}}}, projection, nil)
}

func SearchApps(search string, projection bson.M) (app App, err error) {

	if projection == nil {
		projection = bson.M{}
	}

	var apps []App

	if helpers.RegexNumbers.MatchString(search) {

		id, err := strconv.Atoi(search)
		if err != nil {
			return app, err
		}

		if helpers.IsValidAppID(id) {
			apps, err = GetApps(0, 1, nil, bson.D{{"_id", id}}, projection, nil)
			if err != nil {
				return app, err
			}
		} else {
			return app, ErrInvalidAppID
		}

	} else {
		// filter := bson.D{{"$text", bson.M{"$search": search}}}
		// projection["score"] = bson.M{"$meta": "textScore"}
		// order := bson.D{{"score", bson.M{"$meta": "textScore"}}}
		// apps, err = GetApps(0, 20, order, filter, projection, nil)
		// if err != nil {
		// 	return app, err
		// }
		//
		// if len(apps) == 0 {
		// 	return app, ErrNoDocuments
		// }
		//
		// var names []string
		// for _, v := range apps {
		// 	names = append(names, helpers.RegexNonAlphaNumeric.ReplaceAllString(v.GetName(), ""))
		// }
		// search = helpers.RegexNonAlphaNumeric.ReplaceAllString(search, "")
		// matches := fuzzy.Find(search, names)
		// if len(matches) > 0 {
		// 	return apps[matches[0].Index], nil
		// }

		if !strings.Contains(search, `"`) {
			search = regexp.MustCompile(`([^\s]+)`).ReplaceAllString(search, `"$1"`) // Add quotes to all words
		}

		filter := bson.D{{"$text", bson.M{"$search": search}}}
		apps, err := GetApps(0, 1, bson.D{{"player_peak_week", -1}}, filter, projection, nil)
		if err != nil {
			return app, err
		}
		if len(apps) == 0 {
			return app, ErrNoDocuments
		}

		return apps[0], nil
	}

	return apps[0], nil
}

func GetNonEmptyArrays(offset int64, limit int64, column string, projection bson.M) (apps []App, err error) {

	var filter = bson.D{
		{column, bson.M{"$exists": true}},
		{column, bson.M{"$ne": bson.A{}}},
	}
	var order = bson.D{{"_id", 1}}

	return GetApps(offset, limit, order, filter, projection, nil)
}

func GetRandomApps(count int, filter bson.D, projection bson.M) (apps []App, err error) {

	cur, ctx, err := GetRandomRows(CollectionApps, count, filter, projection)
	if err != nil {
		return apps, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var app App
		err := cur.Decode(&app)
		if err != nil {
			log.Err(err, app.ID)
		}
		apps = append(apps, app)
	}

	return apps, cur.Err()
}

func PopularApps() (apps []App, err error) {

	var item = memcache.MemcachePopularApps

	err = memcache.GetSetInterface(item.Key, item.Expiration, &apps, func() (interface{}, error) {

		return GetApps(
			0,
			30,
			bson.D{{"player_peak_week", -1}},
			bson.D{{"type", "game"}},
			bson.M{"_id": 1, "name": 1, "player_peak_week": 1, "background": 1},
			nil)
	})

	return apps, err
}

func PopularNewApps() (apps []App, err error) {

	var item = memcache.MemcachePopularNewApps

	err = memcache.GetSetInterface(item.Key, item.Expiration, &apps, func() (interface{}, error) {

		releaseDate := time.Now().AddDate(0, 0, -config.Config.NewReleaseDays.GetInt()).Unix()

		return GetApps(
			0,
			12, // Keep even
			bson.D{{Key: "player_peak_week", Value: -1}},
			bson.D{
				{Key: "release_date_unix", Value: bson.M{"$gt": releaseDate}},
				{Key: "type", Value: "game"},
			},
			bson.M{"_id": 1, "name": 1, "player_peak_week": 1},
			nil,
		)
	})

	return apps, err
}

func TrendingApps() (apps []App, err error) {

	var item = memcache.MemcacheTrendingApps

	err = memcache.GetSetInterface(item.Key, item.Expiration, &apps, func() (interface{}, error) {
		return GetApps(
			0,
			10,
			bson.D{{"player_trend", -1}},
			nil,
			bson.M{"_id": 1, "name": 1, "player_trend": 1},
			nil,
		)
	})

	return apps, err
}

func GetAppsGroupedByType(code steamapi.ProductCC) (counts []AppTypeCount, err error) {

	var item = memcache.MemcacheAppTypeCounts(code)

	err = memcache.GetSetInterface(item.Key, item.Expiration, &counts, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return counts, err
		}

		pipeline := mongo.Pipeline{
			{{Key: "$group", Value: bson.M{
				"_id":   "$type",
				"count": bson.M{"$sum": 1},
				"value": bson.M{"$sum": "$prices." + code + ".final"},
			}}},
		}

		cur, err := client.Database(MongoDatabase, options.Database()).Collection(CollectionApps.String()).Aggregate(ctx, pipeline, options.Aggregate())
		if err != nil {
			return counts, err
		}

		defer func() {
			err = cur.Close(ctx)
			log.Err(err)
		}()

		var unknown int64
		var counts []AppTypeCount
		for cur.Next(ctx) {

			var appType AppTypeCount
			err := cur.Decode(&appType)
			if err != nil {
				log.Err(err, appType.Type)
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

	var item = memcache.MemcacheAppReleaseDateCounts

	err = memcache.GetSetInterface(item.Key, item.Expiration, &counts, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return counts, err
		}

		pipeline := mongo.Pipeline{
			bson.D{{Key: "$match", Value: bson.M{"release_date_unix": bson.M{"$gte": time.Now().AddDate(-1, 0, 0).Unix()}}}},
			bson.D{{Key: "$match", Value: bson.M{"release_date_unix": bson.M{"$lte": time.Now().AddDate(0, 0, 1).Unix()}}}},
			bson.D{{Key: "$group", Value: bson.M{"_id": "$release_date_unix", "count": bson.M{"$sum": 1}}}},
		}

		cur, err := client.Database(MongoDatabase, options.Database()).Collection(CollectionApps.String()).Aggregate(ctx, pipeline, options.Aggregate())
		if err != nil {
			return counts, err
		}

		defer func() {
			err = cur.Close(ctx)
			log.Err(err)
		}()

		var counts []AppReleaseDateCount
		for cur.Next(ctx) {

			var row AppReleaseDateCount
			err := cur.Decode(&row)
			if err != nil {
				log.Err(err, row.Date)
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
	Date  int64 `json:"date" bson:"_id"`
	Count int64 `json:"count" bson:"count"`
}

func GetAppsGroupedByReviewScore() (counts []AppReviewScoreCount, err error) {

	var item = memcache.MemcacheAppReviewScoreCounts

	err = memcache.GetSetInterface(item.Key, item.Expiration, &counts, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return counts, err
		}

		pipeline := mongo.Pipeline{
			bson.D{{Key: "$match", Value: bson.M{"reviews_score": bson.M{"$gt": 0}}}},
			bson.D{{Key: "$group", Value: bson.M{"_id": bson.M{"$floor": "$reviews_score"}, "count": bson.M{"$sum": 1}}}},
		}

		cur, err := client.Database(MongoDatabase, options.Database()).Collection(CollectionApps.String()).Aggregate(ctx, pipeline, options.Aggregate())
		if err != nil {
			return counts, err
		}

		defer func() {
			err = cur.Close(ctx)
			log.Err(err)
		}()

		var counts []AppReviewScoreCount
		for cur.Next(ctx) {

			var row AppReviewScoreCount
			err := cur.Decode(&row)
			if err != nil {
				log.Err(err, row.Score)
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

//
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
