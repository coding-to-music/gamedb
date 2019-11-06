package mongo

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
	. "go.mongodb.org/mongo-driver/bson"
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

func (app PackageApp) BSON() D {

	return D{
		{"_id", app.getKey()},
		{"package_id", app.PackageID},
		{"app_id", app.AppID},
		{"app_icon", app.AppIcon},
		{"app_name", app.AppName},
		{"app_type", app.AppType},
		{"app_platforms", app.AppPlatforms},
		{"app_dlc_count", app.AppDLCCount},
	}
}

func (app PackageApp) getKey() string {
	return strconv.Itoa(app.PackageID) + "-" + strconv.Itoa(app.AppID)
}

func UpdatePackageApps(apps []PackageApp) (err error) {

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

	collection := client.Database(MongoDatabase).Collection(CollectionPackageApps.String())
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

	keys := A{}
	for _, appID := range appIDs {

		player := PackageApp{}
		player.PackageID = packageID
		player.AppID = appID

		keys = append(keys, player.getKey())
	}

	collection := client.Database(MongoDatabase).Collection(CollectionPackageApps.String())
	_, err = collection.DeleteMany(ctx, M{"_id": M{"$in": keys}})
	return err
}

func GetPackageApps(packageID int, offset int64, sort D) (apps []PackageApp, err error) {

	var filter = D{{"package_id", packageID}}

	cur, ctx, err := Find(CollectionPackageApps, offset, 100, sort, filter, nil, nil)
	if err != nil {
		return apps, err
	}

	defer func(cur *mongo.Cursor) {
		err = cur.Close(ctx)
		log.Err(err)
	}(cur)

	for cur.Next(ctx) {

		var app PackageApp
		err := cur.Decode(&app)
		if err != nil {
			log.Err(err, app.getKey())
		}
		apps = append(apps, app)
	}

	return apps, cur.Err()
}
