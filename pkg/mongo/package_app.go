package mongo

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PackageApp struct {
	PackageID    int      `bson:"package_id"`
	AppID        int      `bson:"app_id"`
	AppIcon      string   `bson:"app_icon"`
	AppName      string   `bson:"app_name"`
	AppType      string   `bson:"app_type"`
	AppPlatforms []string `bson:"app_platforms"`
	AppDLCCount  int      `bson:"app_dlc_count"`
}

func (app PackageApp) BSON() bson.D {

	return bson.D{
		{Key: "_id", Value: app.getKey()},
		{Key: "package_id", Value: app.PackageID},
		{Key: "app_id", Value: app.AppID},
		{Key: "app_icon", Value: app.AppIcon},
		{Key: "app_name", Value: app.AppName},
		{Key: "app_type", Value: app.AppType},
		{Key: "app_platforms", Value: app.AppPlatforms},
		{Key: "app_dlc_count", Value: app.AppDLCCount},
	}
}

func (app PackageApp) getKey() string {
	return strconv.Itoa(app.PackageID) + "-" + strconv.Itoa(app.AppID)
}

func ReplacePackageApps(apps []PackageApp) (err error) {

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

	collection := client.Database(config.C.MongoDatabase).Collection(CollectionPackageApps.String())
	_, err = collection.BulkWrite(ctx, writes, options.BulkWrite())
	return err
}

func DeletePackageApps(packageID int, appIDs []int) (err error) {

	if len(appIDs) < 1 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	keys := bson.A{}
	for _, appID := range appIDs {

		player := PackageApp{}
		player.PackageID = packageID
		player.AppID = appID

		keys = append(keys, player.getKey())
	}

	collection := client.Database(config.C.MongoDatabase).Collection(CollectionPackageApps.String())
	_, err = collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": keys}})
	return err
}

func GetPackageApps(packageID int, offset int64, sort bson.D) (apps []PackageApp, err error) {

	var filter = bson.D{{"package_id", packageID}}

	cur, ctx, err := find(CollectionPackageApps, offset, 100, filter, sort, nil, nil)
	if err != nil {
		return apps, err
	}

	defer closeCursor(cur, ctx)

	for cur.Next(ctx) {

		var app PackageApp
		err := cur.Decode(&app)
		if err != nil {
			log.ErrS(err, app.getKey())
		} else {
			apps = append(apps, app)
		}
	}

	return apps, cur.Err()
}
