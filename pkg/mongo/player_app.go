package mongo

import (
	"math"
	"strconv"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlayerApp struct {
	PlayerID      int64              `bson:"player_id"`
	PlayerCountry string             `json:"player_country"`
	AppID         int                `bson:"app_id"`
	AppName       string             `bson:"app_name"`
	AppIcon       string             `bson:"app_icon"`
	AppTime       int                `bson:"app_time"`
	AppPrices     map[string]int     `bson:"app_prices"`
	AppPriceHour  map[string]float64 `bson:"app_prices_hour"`

	AppAchievementsTotal   int     `bson:"app_achievements_total"`
	AppAchievementsHave    int     `bson:"app_achievements_have"`
	AppAchievementsPercent float64 `bson:"app_achievements_percent"`
}

func (app PlayerApp) BSON() bson.D {

	var prices = bson.M{}
	for k, v := range app.AppPrices {
		prices[k] = v
	}

	var pricesHour = bson.M{}
	for k, v := range app.AppPriceHour {
		pricesHour[k] = v
	}

	return bson.D{
		{"_id", app.GetKey()},
		{"player_id", app.PlayerID},
		{"app_id", app.AppID},
		{"app_name", app.AppName},
		{"app_icon", app.AppIcon},
		{"app_time", app.AppTime},
		{"app_prices", prices},
		{"app_prices_hour", pricesHour},

		{"app_achievements_total", app.AppAchievementsTotal},
		{"app_achievements_have", app.AppAchievementsHave},
		{"app_achievements_percent", app.AppAchievementsPercent},
	}
}

// Missing achievement columns so when we update a row we dont overwrite achievements
func (app PlayerApp) BSONUpdate() bson.D {

	var prices = bson.M{}
	for k, v := range app.AppPrices {
		prices[k] = v
	}

	var pricesHour = bson.M{}
	for k, v := range app.AppPriceHour {
		pricesHour[k] = v
	}

	return bson.D{
		{"_id", app.GetKey()},
		{"player_id", app.PlayerID},
		{"app_id", app.AppID},
		{"app_name", app.AppName},
		{"app_icon", app.AppIcon},
		{"app_time", app.AppTime},
		{"app_prices", prices},
		{"app_prices_hour", pricesHour},
	}
}

func (app PlayerApp) GetKey() string {
	return strconv.FormatInt(app.PlayerID, 10) + "-" + strconv.Itoa(app.AppID)
}

func (app PlayerApp) GetPath() string {
	return helpers.GetAppPath(app.AppID, app.AppName)
}

func (app PlayerApp) GetIcon() string {
	return helpers.GetAppIcon(app.AppID, app.AppIcon)
}

func (app PlayerApp) GetTimeNice() string {

	return helpers.GetTimeShort(app.AppTime, 2)
}

func (app PlayerApp) GetPriceFormatted(code steamapi.ProductCC) string {

	val, ok := app.AppPrices[string(code)]
	if ok {
		return i18n.FormatPrice(i18n.GetProdCC(code).CurrencyCode, val)
	} else {
		return "-"
	}
}

func (app PlayerApp) GetPriceHourFormatted(code steamapi.ProductCC) string {

	val, ok := app.AppPriceHour[string(code)]
	if ok {
		if val < 0 {
			return "∞"
		}
		return i18n.FormatPrice(i18n.GetProdCC(code).CurrencyCode, int(math.Round(val)))
	} else {
		return "-"
	}
}

func (app PlayerApp) GetAchievementPercent() string {
	return helpers.GetAchievementCompleted(app.AppAchievementsPercent)
}

//noinspection GoUnusedExportedFunction
func CreatePlayerAppIndexes() {

	var indexModels = []mongo.IndexModel{
		{Keys: bson.D{{"app_id", 1}, {"app_time", -1}, {"player_country", 1}}},
		{Keys: bson.D{{"player_id", 1}, {"app_achievements_have", 1}}},
	}

	//
	client, ctx, err := getMongo()
	if err != nil {
		log.Err(err)
		return
	}

	_, err = client.Database(MongoDatabase).Collection(CollectionGroups.String()).Indexes().CreateMany(ctx, indexModels)
	log.Err(err)
}

func GetPlayerAppsByApp(offset int64, filter bson.D) (apps []PlayerApp, err error) {

	return getPlayerApps(offset, 100, filter, bson.D{{"app_time", -1}}, bson.M{"_id": 0, "player_id": 1, "app_time": 1}, nil)
}

func GetPlayerAppsByPlayer(playerID int64, offset int64, limit int64, sort bson.D) (apps []PlayerApp, err error) {

	var filter = bson.D{{"player_id", playerID}}

	return getPlayerApps(offset, limit, filter, sort, bson.M{"app_name": 1, "app_time": 1}, nil)
}

func GetPlayerAppByKey(playerID int64, appID int) (playerApp PlayerApp, err error) {

	playerApp.PlayerID = playerID
	playerApp.AppID = appID

	err = FindOne(CollectionPlayerApps, bson.D{{"_id", playerApp.GetKey()}}, nil, nil, &playerApp)

	return playerApp, err
}

func GetPlayerApps(offset int64, limit int64, filter bson.D, sort bson.D) (apps []PlayerApp, err error) {

	var ops = options.Find().SetHint("player_id_1_app_achievements_have_1")

	return getPlayerApps(offset, limit, filter, sort, nil, ops)
}

func GetPlayersApps(playerIDs []int64, projection bson.M) (apps []PlayerApp, err error) {

	if len(playerIDs) < 1 {
		return apps, err
	}

	playersFilter := bson.A{}
	for _, v := range playerIDs {
		playersFilter = append(playersFilter, v)
	}

	return getPlayerApps(0, 0, bson.D{{"player_id", bson.M{"$in": playersFilter}}}, nil, projection, nil)
}

func GetAppPlayTimes(appID int) ([]PlayerApp, error) {

	return getPlayerApps(0, 0, bson.D{{"app_id", appID}}, nil, bson.M{"_id": 0, "app_time": 1}, nil)
}

func GetAppOwners(appID int) ([]PlayerApp, error) {

	return getPlayerApps(0, 0, bson.D{{"app_id", appID}}, nil, bson.M{"_id": 0, "player_id": 1}, nil)
}

func getPlayerApps(offset int64, limit int64, filter bson.D, sort bson.D, projection bson.M, ops *options.FindOptions) (apps []PlayerApp, err error) {

	cur, ctx, err := Find(CollectionPlayerApps, offset, limit, sort, filter, projection, ops)
	if err != nil {
		return apps, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var playerApp PlayerApp
		err := cur.Decode(&playerApp)
		if err != nil {
			log.Err(err, playerApp.GetKey())
		} else {
			apps = append(apps, playerApp)
		}
	}

	return apps, cur.Err()
}

func UpdatePlayerApps(apps map[int]*PlayerApp) (err error) {

	if len(apps) == 0 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for _, v := range apps {

		// Must be UpdateOneModel, otherwise it will overwrite achievement data
		write := mongo.NewUpdateOneModel()
		write.SetFilter(bson.M{"_id": v.GetKey()})
		write.SetUpdate(bson.M{"$set": v.BSONUpdate()})
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	c := client.Database(MongoDatabase).Collection(CollectionPlayerApps.String())

	_, err = c.BulkWrite(ctx, writes, options.BulkWrite())

	return err
}

func GetAppPlayersByCountry(appID int) (items []PlayerAppsByCountry, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return items, err
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"app_id": appID, "app_time": bson.M{"$ne": 0}}}},
		{{Key: "$group", Value: bson.M{"_id": "$player_country", "count": bson.M{"$sum": 1}}}},
	}

	cur, err := client.Database(MongoDatabase, options.Database()).Collection(CollectionPlayerApps.String()).Aggregate(ctx, pipeline, nil)
	if err != nil {
		return items, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var item PlayerAppsByCountry
		err := cur.Decode(&item)
		if err != nil {
			log.Err(err, item)
		}

		items = append(items, item)
	}

	return items, cur.Err()
}

type PlayerAppsByCountry struct {
	Country string `json:"type" bson:"_id"`
	Count   int64  `json:"count" bson:"count"`
}
