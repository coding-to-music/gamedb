package mongo

import (
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlayerWishlistApp struct {
	PlayerID           int64                      `bson:"player_id"`
	Order              int                        `bson:"order"`
	AppID              int                        `bson:"app_id"`
	AppName            string                     `bson:"app_name"`
	AppIcon            string                     `bson:"app_icon"`
	AppReleaseState    string                     `bson:"app_release_state"`
	AppReleaseDate     time.Time                  `bson:"app_release_date"`
	AppReleaseDateNice string                     `bson:"app_release_date_nice"`
	AppPrices          map[steamapi.ProductCC]int `bson:"app_prices"`
}

func (app PlayerWishlistApp) BSON() bson.D {

	return bson.D{
		{"_id", app.getKey()},
		{"player_id", app.PlayerID},
		{"order", app.Order},
		{"app_id", app.AppID},
		{"app_name", app.AppName},
		{"app_icon", app.AppIcon},
		{"app_release_state", app.AppReleaseState},
		{"app_release_date", app.AppReleaseDate},
		{"app_release_date_nice", app.AppReleaseDateNice},
		{"app_prices", app.AppPrices},
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

func (app PlayerWishlistApp) GetHeaderImage() string {
	return helpers.GetAppHeaderImage(app.AppID)
}

func (app PlayerWishlistApp) GetReleaseState() (ret string) {
	return helpers.GetAppReleaseState(app.AppReleaseState)
}

func (app PlayerWishlistApp) GetReleaseDateNice() string {
	return helpers.GetAppReleaseDateNice(0, app.AppReleaseDate.Unix(), app.AppReleaseDateNice)
}

func ReplacePlayerWishlistApps(apps []PlayerWishlistApp) (err error) {

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
		write.SetFilter(bson.M{"_id": app.getKey()})
		write.SetReplacement(app.BSON())
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	collection := client.Database(config.C.MongoDatabase).Collection(CollectionPlayerWishlistApps.String())
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

	keys := bson.A{}
	for _, appID := range apps {

		player := PlayerWishlistApp{}
		player.PlayerID = playerID
		player.AppID = appID

		keys = append(keys, player.getKey())
	}

	collection := client.Database(config.C.MongoDatabase).Collection(CollectionPlayerWishlistApps.String())
	_, err = collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": keys}})
	return err
}

func GetPlayerWishlistAppsByApp(appID int) (apps []PlayerWishlistApp, err error) {

	return getPlayerWishlistApps(0, 0, bson.D{{"app_id", appID}}, nil, bson.M{"order": 1})
}

func GetPlayerWishlistAppsByPlayer(playerID int64, offset int64, limit int64, order bson.D, projection bson.M) (apps []PlayerWishlistApp, err error) {

	return getPlayerWishlistApps(offset, limit, bson.D{{"player_id", playerID}}, order, projection)
}

func getPlayerWishlistApps(offset int64, limit int64, filter bson.D, sort bson.D, projection bson.M) (apps []PlayerWishlistApp, err error) {

	cur, ctx, err := Find(CollectionPlayerWishlistApps, offset, limit, sort, filter, projection, nil)
	if err != nil {
		return apps, err
	}

	defer closeCursor(cur, ctx)

	for cur.Next(ctx) {

		var app PlayerWishlistApp
		err := cur.Decode(&app)
		if err != nil {
			log.ErrS(err, app.getKey())
		} else {
			apps = append(apps, app)
		}
	}

	return apps, cur.Err()
}
