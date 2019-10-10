package mongo
//
// import (
// 	"github.com/gamedb/gamedb/pkg/helpers"
// 	"github.com/gamedb/gamedb/pkg/log"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )
//
// type WishlistApp struct {
// 	AppID   int    `bson:"_id"`
// 	AppName string `bson:"app_name"`
// 	AppIcon string `bson:"app_icon"`
// 	Count   int    `bson:"count"`
// }
//
// func (wl WishlistApp) BSON() (ret interface{}) {
//
// 	return M{
// 		"_id":      wl.AppID,
// 		"app_name": wl.AppName,
// 		"app_icon": wl.AppIcon,
// 		"count":    wl.Count,
// 	}
// }
//
// func (wl WishlistApp) GetAppIcon() string {
// 	return helpers.GetAppIcon(wl.AppID, wl.AppIcon)
// }
//
// func (wl WishlistApp) GetAppPath() string {
// 	return helpers.GetAppPath(wl.AppID, wl.AppName)
// }
//
// func GetWishlistApps(offset int64) (apps []WishlistApp, err error) {
//
// 	client, ctx, err := getMongo()
// 	if err != nil {
// 		return apps, err
// 	}
//
// 	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionWishlistApps.String())
//
// 	o := options.Find().SetSort(D{{"count", -1}}).SetLimit(100).SetSkip(offset)
//
// 	cur, err := c.Find(ctx, M{}, o)
// 	if err != nil {
// 		return apps, err
// 	}
//
// 	defer func() {
// 		err = cur.Close(ctx)
// 		log.Err(err)
// 	}()
//
// 	for cur.Next(ctx) {
//
// 		var app WishlistApp
// 		err := cur.Decode(&app)
// 		log.Err(err)
// 		apps = append(apps, app)
// 	}
//
// 	return apps, cur.Err()
// }
//
// func UpdateWishlistApps(apps []WishlistApp) (err error) {
//
// 	client, ctx, err := getMongo()
// 	if err != nil {
// 		return err
// 	}
//
// 	collection := client.Database(MongoDatabase).Collection(CollectionWishlistApps.String())
//
// 	var writes []mongo.WriteModel
// 	for _, app := range apps {
//
// 		write := mongo.NewReplaceOneModel()
// 		write.SetFilter(M{"_id": app.AppID})
// 		write.SetReplacement(app.BSON())
// 		write.SetUpsert(true)
//
// 		writes = append(writes, write)
// 	}
//
// 	_, err = collection.BulkWrite(ctx, writes, options.BulkWrite())
// 	log.Err(err)
//
// 	return err
// }
