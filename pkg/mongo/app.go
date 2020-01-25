package mongo

import (
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
	Achievements                  []helpers.AppAchievement       `bson:"achievements"`                    //
	Achievements5                 []helpers.AppAchievement       `bson:"achievements_5"`                  // The first 5 only
	AchievementsAverageCompletion float64                        `bson:"achievements_average_completion"` //
	AchievementsCount             int                            `bson:"achievements_count"`              //
	AlbumMetaData                 pics.AlbumMetaData             `bson:"albummetadata"`                   //
	Background                    string                         `bson:"background"`                      //
	Bundles                       []int                          `bson:"bundle_ids"`                      //
	Categories                    []int                          `bson:"categories"`                      //
	ChangeNumber                  int                            `bson:"change_number"`                   //
	ChangeNumberDate              time.Time                      `bson:"change_number_date"`              //
	ClientIcon                    string                         `bson:"client_icon"`                     //
	ComingSoon                    bool                           `bson:"coming_soon"`                     //
	Common                        map[string]string              `bson:"common"`                          //
	Config                        map[string]string              `bson:"config"`                          //
	CreatedAt                     time.Time                      `bson:"created_at"`                      //
	Demos                         []int                          `bson:"demo_ids"`                        //
	Depots                        pics.Depots                    `bson:"depots"`                          //
	Developers                    []int                          `bson:"developers"`                      //
	DLC                           []int                          `bson:"dlc"`                             //
	DLCCount                      int                            `bson:"dlc_count"`                       //
	Extended                      map[string]string              `bson:"extended"`                        //
	GameID                        int                            `bson:"game_id"`                         //
	GameName                      string                         `bson:"game_name"`                       //
	Genres                        []int                          `bson:"genres"`                          //
	GroupID                       string                         `bson:"group_id"`                        //
	GroupFollowers                int                            `bson:"group_followers"`                 //
	Homepage                      string                         `bson:"homepage"`                        //
	Icon                          string                         `bson:"icon"`                            //
	ID                            int                            `bson:"_id" json:"id"`                   //
	Install                       map[string]interface{}         `bson:"install"`                         //
	IsFree                        bool                           `bson:"is_free"`                         //
	Items                         int                            `bson:"items"`                           //
	ItemsDigest                   string                         `bson:"items_digest"`                    //
	Launch                        []pics.PICSAppConfigLaunchItem `bson:"launch"`                          //
	Localization                  pics.Localisation              `bson:"localization"`                    //
	Logo                          string                         `bson:"logo"`                            //
	MetacriticScore               int8                           `bson:"metacritic_score"`                //
	MetacriticURL                 string                         `bson:"metacritic_url"`                  //
	Movies                        []helpers.AppVideo             `bson:"movies"`                          //
	Name                          string                         `bson:"name"`                            //
	NewsIDs                       []int64                        `bson:"news_ids"`                        //
	Packages                      []int                          `bson:"packages"`                        //
	Platforms                     []string                       `bson:"platforms"`                       //
	PlayerAverageWeek             float64                        `bson:"player_avg_week"`                 //
	PlayerPeakAllTime             int                            `bson:"player_peak_alltime"`             //
	PlayerPeakWeek                int                            `bson:"player_peak_week"`                //
	PlayerTrend                   int64                          `bson:"player_trend"`                    //
	PlaytimeAverage               float64                        `bson:"playtime_average"`                // Minutes
	PlaytimeTotal                 int64                          `bson:"playtime_total"`                  // Minutes
	Prices                        helpers.ProductPrices          `bson:"prices"`                          //
	PublicOnly                    bool                           `bson:"public_only"`                     //
	Publishers                    []int                          `bson:"publishers"`                      //
	RelatedAppIDs                 []int                          `bson:"related_app_ids"`                 //
	RelatedOwnersAppIDs           []helpers.TupleInt             `bson:"related_owners_app_ids"`          //
	ReleaseDate                   string                         `bson:"release_date"`                    //
	ReleaseDateUnix               int64                          `bson:"release_date_unix"`               //
	ReleaseState                  string                         `bson:"release_state"`                   //
	Reviews                       helpers.AppReviewSummary       `bson:"reviews"`                         //
	ReviewsScore                  float64                        `bson:"reviews_score"`                   //
	Screenshots                   []helpers.AppImage             `bson:"screenshots"`                     //
	ShortDescription              string                         `bson:"description_short"`               //
	Stats                         []helpers.AppStat              `bson:"stats"`                           //
	SteamSpy                      helpers.AppSteamSpy            `bson:"steam_spy"`                       //
	SystemRequirements            map[string]interface{}         `bson:"system_requirements"`             //
	Tags                          []int                          `bson:"tags"`                            //
	TwitchID                      int                            `bson:"twitch_id"`                       //
	TwitchURL                     string                         `bson:"twitch_url"`                      //
	Type                          string                         `bson:"type"`                            //
	UFS                           map[string]string              `bson:"ufs"`                             //
	UpdatedAt                     time.Time                      `bson:"updated_at"`                      //
	Version                       string                         `bson:"version"`                         //
	WishlistAvgPosition           float64                        `bson:"wishlist_avg_position"`           //
	WishlistCount                 int                            `bson:"wishlist_count"`                  //
}

func (app App) BSON() bson.D {

	// Set Achievements5
	app.Achievements5 = []helpers.AppAchievement{}
	for _, v := range app.Achievements {
		if v.Active && strings.HasSuffix(v.Icon, ".jpg") {
			if len(app.Achievements5) < 5 {
				app.Achievements5 = append(app.Achievements5, v)
			} else {
				break
			}
		}
	}

	if app.ChangeNumberDate.IsZero() {
		app.ChangeNumberDate = time.Now()
	}

	app.UpdatedAt = time.Now()

	return bson.D{
		{"achievements", app.Achievements},
		{"achievements_5", app.Achievements5},
		{"achievements_average_completion", app.AchievementsAverageCompletion},
		{"achievements_count", app.AchievementsCount},
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

func (app App) GetID() int {
	return app.ID
}

func (app App) GetProductType() helpers.ProductType {
	return helpers.ProductTypeApp
}

func (app App) GetPrice(code steam.ProductCC) (price helpers.ProductPrice) {
	return app.Prices.Get(code)
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

	if helpers.SliceHasString(app.Platforms, platformWindows) {
		ret = ret + `<a href="/apps?platforms=windows"><i class="fab fa-windows" data-toggle="tooltip" data-placement="top" title="Windows"></i></a>`
	} else {
		ret = ret + `<span class="space"></span>`
	}

	if helpers.SliceHasString(app.Platforms, platformMac) {
		ret = ret + `<a href="/apps?platforms=macos"><i class="fab fa-apple" data-toggle="tooltip" data-placement="top" title="Mac"></i></a>`
	} else {
		ret = ret + `<span class="space"></span>`
	}

	if helpers.SliceHasString(app.Platforms, platformLinux) {
		ret = ret + `<a href="/apps?platforms=linux"><i class="fab fa-linux" data-toggle="tooltip" data-placement="top" title="Linux"></i></a>`
	} else {
		ret = ret + `<span class="space"></span>`
	}

	return ret, nil
}

func (app App) GetMetaImage() string {

	ss := app.Screenshots
	if len(ss) == 0 {
		return app.GetHeaderImage()
	}
	return ss[0].PathFull
}

func (app App) GetAppRelatedApps() (apps []App, err error) {

	apps = []App{} // Needed for marshalling into type

	if len(app.RelatedAppIDs) == 0 {
		return apps, nil
	}

	var item = memcache.MemcacheAppRelated(app.ID)

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &apps, func() (interface{}, error) {

		return GetAppsByID(app.RelatedAppIDs, bson.M{"_id": 1, "name": 1})
	})

	return apps, err
}

func (app App) GetDemos() (demos []App, err error) {

	demos = []App{} // Needed for marshalling into type

	if len(app.Demos) == 0 {
		return demos, nil
	}

	var item = memcache.MemcacheAppDemos(app.ID)

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &demos, func() (interface{}, error) {
		return GetAppsByID(app.Demos, bson.M{"_id": 1, "name": 1})
	})

	return demos, err
}

func (app App) GetDLCs() (apps []App, err error) {

	apps = []App{} // Needed for marshalling into type

	if len(app.DLC) == 0 {
		return apps, nil
	}

	var item = memcache.MemcacheAppDLC(app.ID)

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &apps, func() (interface{}, error) {

		return GetAppsByID(app.DLC, bson.M{"_id": 1, "name": 1})
	})

	return apps, err
}

func (app App) ReadPICS(m map[string]string) (config pics.PICSKeyValues) {

	config = pics.PICSKeyValues{}

	for k, v := range m {
		config[k] = v
	}

	return config
}

func (app App) GetReleaseDateNice() string {
	return helpers.GetAppReleaseDateNice(app.ReleaseDateUnix, app.ReleaseDate)
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

func (app App) GetUpdatedNice() string {
	return app.UpdatedAt.Format(helpers.DateYearTime)
}

func (app App) Save() (err error) {

	if app.ID == 0 {
		return errors.New("invalid app id")
	}

	_, err = ReplaceOne(CollectionApps, bson.D{{"_id", app.ID}}, app)
	return err
}

func CreateAppIndexes() {

	var ascending = []string{
		"achievements_average_completion",
		"achievements_count",
		"categories",
		"developers",
		"genres",
		"group_followers",
		"platforms",
		"player_peak_week",
		"player_trend",
		// "prices",
		"publishers",
		"release_date_unix",
		"reviews_score",
		"tags",
		"type",
		"wishlist_avg_position",
	}

	var descending = []string{
		"achievements_average_completion",
		"achievements_count",
		"group_followers",
		"player_peak_week",
		"player_trend",
		// "prices",
		"release_date_unix",
		"reviews_score",
		"wishlist_avg_position",
		"wishlist_count",
	}

	// Price fields
	for _, v := range helpers.GetProdCCs(true) {
		ascending = append(ascending, "prices."+string(v.ProductCode)+".final")
		descending = append(descending, "prices."+string(v.ProductCode)+".final")
	}

	//
	var indexModels []mongo.IndexModel
	for _, v := range ascending {
		indexModels = append(indexModels, mongo.IndexModel{
			Keys: bson.D{{v, 1}},
		})
	}

	for _, v := range descending {
		indexModels = append(indexModels, mongo.IndexModel{
			Keys: bson.D{{v, -1}},
		})
	}

	// Text index
	indexModels = append(indexModels, mongo.IndexModel{
		Keys:    bson.D{{"name", "text"}},
		Options: options.Index().SetName("text"),
	})

	// Achievements page
	indexModels = append(indexModels, mongo.IndexModel{
		Keys: bson.D{{"achievements_count", 1}},
	})
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

func GetApp(id int, projection bson.M) (app App, err error) {

	if !helpers.IsValidAppID(id) {
		return app, ErrInvalidAppID
	}

	if id == 0 {
		id = 753
	}

	err = FindOne(CollectionApps, bson.D{{"_id", id}}, nil, projection, &app)
	if err != nil {
		return app, err
	}
	if app.ID == 0 {
		return app, ErrNoDocuments
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

	a := bson.A{}
	for _, v := range ids {
		a = append(a, v)
	}

	return GetApps(0, 0, nil, bson.D{{"_id", bson.M{"$in": a}}}, projection, nil)
}

func SearchApps(search string, projection bson.M) (app App, err error) {

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
		filter := bson.D{{"$text", bson.M{"$search": search}}}
		projection["score"] = bson.M{"$meta": "textScore"}
		order := bson.D{{"score", bson.M{"$meta": "textScore"}}}
		apps, err = GetApps(0, 1, order, filter, projection, nil)
		if err != nil {
			return app, err
		}
	}

	if len(apps) == 0 {
		return app, ErrNoDocuments
	}

	return apps[0], nil
}

func GetNonEmptyArrays(column string, projection bson.M) (apps []App, err error) {

	var filter = bson.D{{column + ".0", bson.M{"$exists": true}}}
	var order = bson.D{{"_id", 1}}

	return GetApps(0, 0, order, filter, projection, nil)
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

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &apps, func() (interface{}, error) {

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

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &apps, func() (interface{}, error) {

		releaseDate := time.Now().AddDate(0, 0, -config.Config.NewReleaseDays.GetInt()).Unix()

		return GetApps(
			0,
			25,
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

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &apps, func() (interface{}, error) {
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

// func WishlistedApps() (appsMap map[int]bool, err error) {
//
// 	apps, err:= GetApps(0)
//
// 	db, err := GetMySQLClient()
// 	if err != nil {
// 		return appsMap, err
// 	}
//
// 	var apps []App
// 	db = db.Select([]string{"id"})
// 	db = db.Where("wishlist_count > ?", 0)
// 	db = db.Find(&apps)
//
// 	appsMap = map[int]bool{}
// 	for _, app := range apps {
// 		appsMap[app.ID] = true
// 	}
//
// 	return appsMap, err
// }

func GetAppTypes() (counts []AppTypeCount, err error) {

	var item = memcache.MemcacheAppTypesCounts

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &counts, func() (interface{}, error) {

		client, ctx, err := getMongo()
		if err != nil {
			return counts, err
		}

		pipeline := mongo.Pipeline{
			{{Key: "$group", Value: bson.M{"_id": "$type", "count": bson.M{"$sum": 1}}}},
		}

		cur, err := client.Database(MongoDatabase, options.Database()).Collection(CollectionApps.String()).Aggregate(ctx, pipeline, options.Aggregate())
		if err != nil {
			return counts, err
		}

		defer func() {
			err = cur.Close(ctx)
			log.Err(err)
		}()

		var unknown int
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
	Count int    `json:"count"`
}

func (atc AppTypeCount) Format() string {
	return helpers.GetAppType(atc.Type)
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
