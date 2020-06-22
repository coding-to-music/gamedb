package queue

import (
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	steamHelper "github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type PlayerAchievementsMessage struct {
	PlayerID int64 `json:"player_id"`
	AppID    int   `json:"app_id"`
	Force    bool  `json:"force"`
}

var appsWithNoStats = map[int]bool{}

func playerAchievementsHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := PlayerAchievementsMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		if _, ok := appsWithNoStats[payload.AppID]; ok {
			message.Ack(false)
			continue
		}

		// Get app
		app, err := mongo.GetApp(payload.AppID, false)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToRetryQueue(message)
			continue
		}

		if app.AchievementsCount == 0 {
			appsWithNoStats[payload.AppID] = true
		}

		// Get player
		player, err := mongo.GetPlayer(payload.PlayerID)
		if err != nil {

			// ErrNoDocuments can be returned on new signups as the player hasnt been created yet
			err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
			if err != nil {
				log.Err(err, message.Message.Body)
			}

			sendToRetryQueueWithDelay(message, time.Second*10)
			continue
		}

		// Do API call
		resp, _, err := steamHelper.GetSteamUnlimited().GetPlayerAchievements(uint64(payload.PlayerID), uint32(payload.AppID))
		err = steamHelper.AllowSteamCodes(err, 400)
		if err != nil {
			steamHelper.LogSteamError(err, message.Message.Body)
			sendToRetryQueue(message)
			continue
		}

		if !resp.Success {

			if resp.Error == "Requested app has no stats" {
				appsWithNoStats[payload.AppID] = true
			}

			message.Ack(false)
			continue
		}

		// Get the last saved achievement
		var timestamp int64
		if !config.IsLocal() && !payload.Force {
			timestamp, err = mongo.FindLatestPlayerAchievement(payload.PlayerID, payload.AppID)
			if err != nil {
				log.Err(err)
				sendToRetryQueue(message)
				continue
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
				log.Err(err)
				sendToRetryQueue(message)
				continue
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

		err = mongo.UpdatePlayerAchievements(rows)
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			continue
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

		_, err = mongo.UpdateOne(mongo.CollectionPlayerApps, bson.D{{"_id", playerApp.GetKey()}}, bson.D{
			{"app_achievements_total", app.AchievementsCount},
			{"app_achievements_have", have},
			{"app_achievements_percent", percent},
		})
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
