package queue

import (
	"sort"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/steam"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type AppAchievementsMessage struct {
	AppID     int    `json:"id"`
	AppName   string `json:"app-name"`
	AppOwners int64  `json:"app_owners"`
}

func (m AppAchievementsMessage) Queue() rabbit.QueueName {
	return QueueAppsAchievements
}

func appAchievementsHandler(message *rabbit.Message) {

	payload := AppAchievementsMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	//
	schemaResponse, err := steam.GetSteam().GetSchemaForGame(payload.AppID)
	err = steam.AllowSteamCodes(err, 400, 403)
	if err != nil {
		steam.LogSteamError(err)
		sendToRetryQueue(message)
		return
	}

	globalResponse, err := steam.GetSteam().GetGlobalAchievementPercentagesForApp(payload.AppID)
	err = steam.AllowSteamCodes(err, 403, 500)
	if err != nil {
		steam.LogSteamError(err)
		sendToRetryQueue(message)
		return
	}

	// Build map of all global achievements
	var achievementsMap = map[string]mongo.AppAchievement{}

	for _, achievement := range globalResponse.GlobalAchievementPercentage {

		achievementsMap[achievement.Name] = mongo.AppAchievement{
			AppID:     payload.AppID,
			Key:       achievement.Name,
			Completed: achievement.Percent,
			Deleted:   false,
		}
	}

	// Add in data for achievements in schema
	var percentTotal float64
	var percentCount int

	for _, achievement := range schemaResponse.AvailableGameStats.Achievements {

		if val, ok := achievementsMap[achievement.Name]; ok {

			percentTotal += val.Completed
			percentCount++

			val.Fill(payload.AppID, achievement)
			val.Active = true

			achievementsMap[achievement.Name] = val

		} else {

			val := mongo.AppAchievement{}
			val.Fill(payload.AppID, achievement)

			achievementsMap[achievement.Name] = val
		}
	}

	// Update player achievements, too much cpu
	// for _, achievement := range achievementsMap {
	//
	// 	var filter = bson.D{
	// 		{"app_id", achievement.AppID},
	// 		{"achievement_id", achievement.Key},
	// 		{"achievement_complete", bson.M{"$ne": achievement.Completed}},
	// 	}
	//
	// 	var update = bson.D{
	// 		{"achievement_complete", achievement.Completed},
	// 	}
	//
	// 	_, err = mongo.UpdateManySet(mongo.CollectionPlayerAchievements, filter, update)
	// 	if err != nil {
	// 		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
	// 	}
	// }

	// Save achievements to Mongo
	var achievementsSlice []mongo.AppAchievement
	for _, achievement := range achievementsMap {
		achievementsSlice = append(achievementsSlice, achievement)
	}

	// Sort by key to store the first 5
	sort.Slice(achievementsSlice, func(i, j int) bool {
		return achievementsSlice[i].Completed > achievementsSlice[j].Completed
	})

	err = mongo.ReplaceAppAchievements(achievementsSlice)
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	// Update in Elastic
	for _, v := range achievementsSlice {
		err = ProduceAchievementSearch(v, payload.AppName, payload.AppOwners)
		if err != nil {
			log.ErrS(err)
		}
	}

	// Update app row
	var achievementsCol []helpers.Tuple
	for _, achievement := range achievementsSlice {
		if achievement.Active && achievement.Icon != "" {
			if len(achievementsCol) < 5 {
				achievementsCol = append(achievementsCol, helpers.Tuple{
					Key:   achievement.Icon,
					Value: achievement.Name,
				})
			} else {
				break
			}
		}
	}

	var average float64
	if percentCount != 0 {
		average = percentTotal / float64(percentCount)
	}

	var stats []helpers.AppStat
	for _, v := range schemaResponse.AvailableGameStats.Stats {
		stats = append(stats, helpers.AppStat{
			Name:        v.Name,
			Default:     v.DefaultValue,
			DisplayName: v.DisplayName,
		})
	}

	var updateApp = bson.D{
		{"version", schemaResponse.Version},
		{"achievements_count", len(schemaResponse.AvailableGameStats.Achievements)},
		{"achievements_count_total", len(globalResponse.GlobalAchievementPercentage)},
		{"achievements_5", achievementsCol},
		{"achievements_average_completion", average},
		{"stats", stats},
	}

	_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{"_id", payload.AppID}}, updateApp)
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	// Mark apps in Mongo but not in global response as deleted
	var filter = bson.D{{"app_id", payload.AppID}}
	if len(achievementsMap) > 0 {
		var keys []string
		for k := range achievementsMap {
			keys = append(keys, k)
		}
		filter = append(filter, bson.E{Key: "key", Value: bson.M{"$nin": keys}})
	}

	_, err = mongo.UpdateManySet(mongo.CollectionAppAchievements, filter, bson.D{{"deleted", true}})
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	//
	var items = []string{
		memcache.ItemApp(payload.AppID).Key,
		memcache.ItemMongoCount(mongo.CollectionAppAchievements.String(), bson.D{{"app_id", payload.AppID}}).Key,
	}

	err = memcache.Delete(items...)
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	// Update in Elastic
	err = ProduceAppSearch(nil, payload.AppID)
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	message.Ack()
}
