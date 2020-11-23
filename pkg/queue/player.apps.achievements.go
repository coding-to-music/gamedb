package queue

import (
	"strconv"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/steam"
	"github.com/gamedb/gamedb/pkg/websockets"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type PlayerAchievementsMessage struct {
	PlayerID     int64 `json:"player_id"`
	AppID        int   `json:"app_id"`
	Force        bool  `json:"force"` // Re-add previous achievements
	OldCount     int   `json:"old_count"`
	OldCount100  int   `json:"old_count_100"`
	OldCountApps int   `json:"old_count_apps"`
}

func playerAchievementsHandler(message *rabbit.Message) {

	payload := PlayerAchievementsMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	// Last item for this player
	if payload.AppID == 0 {

		<-time.NewTimer(time.Second * 10).C // Sleep to make sure all other messages are consumed

		// Websocket
		defer func() {

			wsPayload := PlayerPayload{
				ID:    strconv.FormatInt(payload.PlayerID, 10),
				Queue: "achievement",
			}

			err = ProduceWebsocket(wsPayload, websockets.PagePlayer)
			if err != nil {
				log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
			}
		}()

		// Total achievements
		count, err := mongo.CountDocuments(mongo.CollectionPlayerAchievements, bson.D{{"player_id", payload.PlayerID}}, 0)
		if err != nil {
			log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
			sendToRetryQueue(message)
			return
		}

		count100, err := mongo.CountDocuments(mongo.CollectionPlayerApps, bson.D{{"player_id", payload.PlayerID}, {"app_achievements_percent", 100}}, 0)
		if err != nil {
			log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
			sendToRetryQueue(message)
			return
		}

		countApps, err := mongo.CountDocuments(mongo.CollectionPlayerApps, bson.D{{"player_id", payload.PlayerID}, {"app_achievements_have", bson.M{"$gt": 0}}}, 0)
		if err != nil {
			log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
			sendToRetryQueue(message)
			return
		}

		if payload.OldCount > int(count) {
			count = int64(payload.OldCount)
		}
		if payload.OldCount100 > int(count100) {
			count100 = int64(payload.OldCount100)
		}
		if payload.OldCountApps > int(countApps) {
			countApps = int64(payload.OldCountApps)
		}

		// Update Mongo
		var update = bson.D{
			{"achievement_count", int(count)},
			{"achievement_count_100", int(count100)},
			{"achievement_count_apps", int(countApps)},
		}

		_, err = mongo.UpdateOne(mongo.CollectionPlayers, bson.D{{"_id", payload.PlayerID}}, update)
		if err != nil {
			log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
			sendToRetryQueue(message)
			return
		}

		// Update Influx
		err = savePlayerStatsToInflux(payload.PlayerID, map[string]interface{}{
			influx.InfPlayersAchievements.String(): count,
		})
		if err != nil {
			log.ErrS(err, payload.PlayerID)
			sendToRetryQueue(message)
			return
		}

		// Clear caches
		var items = []string{
			memcache.MemcachePlayer(payload.PlayerID).Key,
			memcache.MemcacheMongoCount(mongo.CollectionPlayerAchievements.String(), bson.D{{"player_id", payload.PlayerID}}).Key,
		}

		err = memcache.Delete(items...)
		if err != nil {
			log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
			sendToRetryQueue(message)
			return
		}

		// Update Elastic
		err = ProducePlayerSearch(nil, payload.PlayerID)
		if err != nil {
			log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
			sendToRetryQueue(message)
			return
		}

		//
		message.Ack()
		return
	}

	item := memcache.MemcacheAppNoAchievements(payload.AppID) // Cache if this app has no achievements

	_, err = memcache.Get(item.Key)
	if err == nil {
		message.Ack()
		return
	}

	// Get app
	app, err := mongo.GetApp(payload.AppID, false)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	if app.AchievementsCount == 0 {
		err = memcache.Set(item.Key, item.Value, item.Expiration)
		if err != nil {
			log.ErrS(err)
		}
	}

	// Get player
	player, err := mongo.GetPlayer(payload.PlayerID)
	if err != nil {

		// ErrNoDocuments can be returned on new signups as the player hasnt been created yet
		err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
		if err != nil {
			log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		}

		sendToRetryQueueWithDelay(message, time.Second*10)
		return
	}

	// Do API call
	resp, err := steam.GetSteamUnlimited().GetPlayerAchievements(uint64(payload.PlayerID), uint32(payload.AppID))

	// Skip private profiles
	if val, ok := err.(steamapi.Error); ok && val.Code == 403 {
		message.Ack()
		return
	}

	err = steam.AllowSteamCodes(err, 400)
	if err != nil {
		steam.LogSteamError(err, string(message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	if !resp.Success {

		if resp.Error == "Requested app has no stats" {
			err = memcache.Set(item.Key, item.Value, item.Expiration)
			if err != nil {
				log.ErrS(err)
			}
		}

		message.Ack()
		return
	}

	// Get the last saved achievement
	var timestamp int64
	if !config.IsLocal() && !payload.Force {
		timestamp, err = mongo.FindLatestPlayerAchievement(payload.PlayerID, payload.AppID)
		if err != nil {
			log.ErrS(err)
			sendToRetryQueue(message)
			return
		}
	}

	// Get achievements for icons
	var a bson.A
	for _, v := range resp.Achievements {
		if v.Achieved && v.UnlockTime >= timestamp {
			a = append(a, v.APIName)
		}
	}

	var appAchievementsMap = map[string]mongo.AppAchievement{}

	if len(a) > 0 {

		var filter = bson.D{
			{"app_id", payload.AppID},
			{"key", bson.M{"$in": a}},
		}

		appAchievements, err := mongo.GetAppAchievements(0, 0, filter, nil)
		if err != nil {
			log.ErrS(err)
			sendToRetryQueue(message)
			return
		}

		for _, appAchievement := range appAchievements {
			appAchievementsMap[appAchievement.Key] = appAchievement
		}
	}

	// Save new player achievements
	var rows []mongo.PlayerAchievement

	for _, v := range resp.Achievements {
		if v.Achieved && v.UnlockTime >= timestamp {

			appAchievement, _ := appAchievementsMap[v.APIName]

			rows = append(rows, mongo.PlayerAchievement{
				PlayerID:               payload.PlayerID,
				PlayerName:             player.PersonaName,
				PlayerIcon:             player.Avatar,
				AppID:                  app.ID,
				AppName:                app.Name,
				AppIcon:                app.Icon,
				AchievementID:          v.APIName,
				AchievementName:        v.Name,
				AchievementDescription: v.Description,
				AchievementDate:        v.UnlockTime,
				AchievementIcon:        appAchievement.Icon,
				AchievementComplete:    appAchievement.Completed,
			})
		}
	}

	err = mongo.ReplacePlayerAchievements(rows)
	if err != nil {
		log.ErrS(err)
		sendToRetryQueue(message)
		return
	}

	// Update player_apps row
	playerApp := mongo.PlayerApp{}
	playerApp.PlayerID = payload.PlayerID
	playerApp.AppID = payload.AppID

	var have int
	for _, v := range resp.Achievements {
		if v.Achieved {
			have++
		}
	}

	var percent float64
	if have > 0 && app.AchievementsCount > 0 {
		percent = float64(have) / float64(app.AchievementsCount) * 100
	}

	var update = bson.D{
		{"app_achievements_total", app.AchievementsCount},
		{"app_achievements_have", have},
		{"app_achievements_percent", percent},
	}

	_, err = mongo.UpdateOne(mongo.CollectionPlayerApps, bson.D{{"_id", playerApp.GetKey()}}, update)
	if err != nil {
		log.ErrS(err)
		sendToRetryQueue(message)
		return
	}

	message.Ack()
}
