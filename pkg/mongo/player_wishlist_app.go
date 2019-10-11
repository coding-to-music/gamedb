package mongo

import (
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlayerWishlistApp struct {
	PlayerID           int64                   `bson:"player_id"`
	Order              int                     `bson:"order"`
	AppID              int                     `bson:"app_id"`
	AppName            string                  `bson:"app_name"`
	AppIcon            string                  `bson:"app_icon"`
	AppReleaseState    string                  `bson:"app_release_state"`
	AppReleaseDate     time.Time               `bson:"app_release_date"`
	AppReleaseDateNice string                  `bson:"app_release_date_nice"`
	AppPrices          map[steam.ProductCC]int `bson:"app_prices"`
}

func (app PlayerWishlistApp) BSON() (ret interface{}) {

	return M{
		"_id":                   app.getKey(),
		"player_id":             app.PlayerID,
		"order":                 app.Order,
		"app_id":                app.AppID,
		"app_name":              app.AppName,
		"app_icon":              app.AppIcon,
		"app_release_state":     app.AppReleaseState,
		"app_release_date":      app.AppReleaseDate,
		"app_release_date_nice": app.AppReleaseDateNice,
		"app_prices":            app.AppPrices,
	}
}

func (app PlayerWishlistApp) getKey() string {
	return strconv.FormatInt(app.PlayerID, 10) + "-" + strconv.Itoa(app.AppID)
}

func (app PlayerWishlistApp) GetName() string {
	return helpers.GetAppName(app.AppID, app.AppName)
}

func (app PlayerWishlistApp) GetPath() string {
	return helpers.GetAppPath(app.AppID, app.AppName)
}

func (app PlayerWishlistApp) GetIcon() (ret string) {
	return helpers.GetAppIcon(app.AppID, app.AppIcon)
}

func (app PlayerWishlistApp) GetReleaseState() (ret string) {
	return helpers.GetAppReleaseState(app.AppReleaseState)
}

func (app PlayerWishlistApp) GetReleaseDateNice() string {
	return helpers.GetAppReleaseDateNice(app.AppReleaseDate.Unix(), app.AppReleaseDateNice)
}

func InsertPlayerWishlistApps(apps []PlayerWishlistApp) (err error) {

	if len(apps) < 1 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for _, app := range apps {

		write := mongo.NewReplaceOneModel()
		write.SetFilter(M{"_id": app.getKey()})
		write.SetReplacement(app.BSON())
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	collection := client.Database(MongoDatabase).Collection(CollectionPlayerWishlistApps.String())
	_, err = collection.BulkWrite(ctx, writes, options.BulkWrite())
	return err
}

func DeletePlayerWishlistApps(playerID int64, apps []int) (err error) {

	if len(apps) < 1 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	keys := A{}
	for _, appID := range apps {

		player := PlayerWishlistApp{}
		player.PlayerID = playerID
		player.AppID = appID

		keys = append(keys, player.getKey())
	}

	collection := client.Database(MongoDatabase).Collection(CollectionPlayerWishlistApps.String())
	_, err = collection.DeleteMany(ctx, M{"_id": M{"$in": keys}})
	return err
}

// Used when app gets updated
func GetPlayerWishlistAppsByApp(appID int) (apps []PlayerWishlistApp, err error) {

	return getPlayerWishlistApps(0, 0, M{"app_id": appID}, nil)
}

// Used to show player groups on frontend
func GetPlayerWishlistAppsByPlayer(playerID int64, offset int64, order D) (apps []PlayerWishlistApp, err error) {

	return getPlayerWishlistApps(offset, 100, M{"player_id": playerID}, order)
}

// Used when player is updated
func GetAllPlayerWishlistApps(playerID int64) (apps []PlayerWishlistApp, err error) {

	return getPlayerWishlistApps(0, 0, M{"player_id": playerID}, nil)
}

func getPlayerWishlistApps(offset int64, limit int64, filter M, sort D) (apps []PlayerWishlistApp, err error) {

	if filter == nil {
		filter = M{}
	}

	client, ctx, err := getMongo()
	if err != nil {
		return apps, err
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionPlayerWishlistApps.String())

	ops := options.Find()

	if sort != nil {
		ops.SetSort(sort)
	}
	if limit > 0 {
		ops.SetLimit(limit)
	}
	if offset > 0 {
		ops.SetSkip(offset)
	}

	cur, err := c.Find(ctx, filter, ops)
	if err != nil {
		return apps, err
	}

	defer func(cur *mongo.Cursor) {
		err = cur.Close(ctx)
		log.Err(err)
	}(cur)

	for cur.Next(ctx) {

		var app PlayerWishlistApp
		err := cur.Decode(&app)
		if err != nil {
			log.Err(err, app.getKey())
		}
		apps = append(apps, app)
	}

	return apps, cur.Err()
}
