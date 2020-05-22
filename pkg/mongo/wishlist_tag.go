package mongo

//
// import (
// 	"strconv"
//
// 	"github.com/gamedb/gamedb/pkg/log"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )
//
// type WishlistTag struct {
// 	TagID   int    `bson:"_id"`
// 	TagName string `bson:"tag_name"`
// 	Count   int    `bson:"count"`
// }
//
// func (wl WishlistTag) BSON() (ret interface{}) {
//
// 	return M{
// 		"_id":      wl.TagID,
// 		"tag_name": wl.TagName,
// 		"count":    wl.Count,
// 	}
// }
//
// func (wl WishlistTag) GetTagPath() string {
// 	return "/games?tags=" + strconv.Itoa(wl.TagID)
// }
//
// func GetWishlistTags() (tags []WishlistTag, err error) {
//
// 	client, ctx, err := getMongo()
// 	if err != nil {
// 		return tags, err
// 	}
//
// 	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionWishlistTags.String())
//
// 	o := options.Find().SetSort(D{{"count", -1}})
//
// 	cur, err := c.Find(ctx, M{}, o)
// 	if err != nil {
// 		return tags, err
// 	}
//
// 	defer func() {
// 		err = cur.Close(ctx)
// 		log.Err(err)
// 	}()
//
// 	for cur.Next(ctx) {
//
// 		var tag WishlistTag
// 		err := cur.Decode(&tag)
// 		log.Err(err)
// 		tags = append(tags, tag)
// 	}
//
// 	return tags, cur.Err()
// }
//
// func UpdateWishlistTags(apps []WishlistTag) (err error) {
//
// 	client, ctx, err := getMongo()
// 	if err != nil {
// 		return err
// 	}
//
// 	collection := client.Database(MongoDatabase).Collection(CollectionWishlistTags.String())
//
// 	var writes []mongo.WriteModel
// 	for _, app := range apps {
//
// 		write := mongo.NewReplaceOneModel()
// 		write.SetFilter(M{"_id": app.TagID})
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
