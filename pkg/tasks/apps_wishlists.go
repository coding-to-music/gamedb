package tasks

//
// import (
// 	"strconv"
//
// 	"github.com/gamedb/gamedb/pkg/log"
// 	"github.com/gamedb/gamedb/pkg/mongo"
// 	"github.com/gamedb/gamedb/pkg/sql"
// 	"go.mongodb.org/mongo-driver/mongo/options"
//  . "go.mongodb.org/mongo-driver/bson"
// )
//
// type Wishlists struct {
// 	BaseTask
// }
//
// func (c Wishlists) ID() string {
// 	return "update-wishlists"
// }
//
// func (c Wishlists) Name() string {
// 	return "Update wishlist"
// }
//
// func (c Wishlists) Cron() string {
// 	return CronTimeWishlists
// }
//
// func (c Wishlists) work() (err error) {
//
// 	// Get app counts
// 	var appCounts = map[int]int{}
//
// 	ops := options.Find().SetReturnKey(false)
// 	players, err := mongo.GetPlayers(0, 0, nil, M{"wishlist_app_ids": M{"$exists": true, "$not": M{"$size": 0}}}, M{"_id": 1, "wishlist_app_ids": 1}, ops)
// 	if err != nil {
// 		log.Err(err)
// 		return
// 	}
//
// 	log.Info("Found " + strconv.Itoa(len(players)) + " players")
//
// 	for _, v := range players {
// 		for _, vv := range v.Wishlist {
// 			appCounts[vv]++
// 		}
// 	}
//
// 	// Get apps
// 	var appIDs []int
// 	for appID := range appCounts {
// 		appIDs = append(appIDs, appID)
// 	}
//
// 	apps, err := sql.GetAppsByID(appIDs, []string{"id", "name", "tags", "icon"})
// 	if err != nil {
// 		log.Err(err)
// 		return
// 	}
//
// 	log.Info("Found " + strconv.Itoa(len(apps)) + " apps")
//
// 	var appMap = map[int]sql.App{}
// 	for _, app := range apps {
// 		appMap[app.ID] = app
// 	}
//
// 	// Get tag counts
// 	var tagCounts = map[int]int{}
//
// 	for _, player := range players {
// 		for _, appID := range player.Wishlist {
//
// 			val, ok := appMap[appID]
// 			if ok {
// 				tags, err := val.GetTags()
// 				if err != nil {
// 					log.Err(err)
// 					continue
// 				}
// 				for _, tag := range tags {
// 					tagCounts[tag.ID]++
// 				}
// 			}
// 		}
// 	}
//
// 	// Get tags
// 	tags, err := sql.GetAllTags()
// 	if err != nil {
// 		log.Err(err)
// 		return
// 	}
//
// 	log.Info("Found " + strconv.Itoa(len(tags)) + " tags")
//
// 	var tagMap = map[int]sql.Tag{}
// 	for _, tag := range tags {
// 		tagMap[tag.ID] = tag
// 	}
//
// 	//
// 	var wishlistApps []mongo.WishlistApp
// 	for appID, count := range appCounts {
//
// 		wishlistApps = append(wishlistApps, mongo.WishlistApp{
// 			AppID:   appID,
// 			AppName: appMap[appID].GetName(),
// 			Count:   count,
// 			AppIcon: appMap[appID].Icon,
// 		})
// 	}
//
// 	var wishlistTags []mongo.WishlistTag
// 	for tagID, count := range tagCounts {
//
// 		wishlistTags = append(wishlistTags, mongo.WishlistTag{
// 			TagID:   tagID,
// 			TagName: tagMap[tagID].GetName(),
// 			Count:   count,
// 		})
// 	}
//
// 	//
// 	err = mongo.UpdateWishlistApps(wishlistApps)
// 	log.Err(err)
//
// 	err = mongo.UpdateWishlistTags(wishlistTags)
// 	log.Err(err)
// }
