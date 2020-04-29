package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	steamHelper "github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type PlayerAchievementsMessage struct {
	PlayerID int64 `json:"player_id"`
	AppID    int   `json:"app_id"`
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

		app, err := mongo.GetApp(payload.AppID, false)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToRetryQueue(message)
			continue
		}

		if app.AchievementsCountTotal == 0 {
			appsWithNoStats[payload.AppID] = true
		}

		resp, b, err := steamHelper.GetSteamUnlimited().GetPlayerAchievements(uint64(payload.PlayerID), uint32(payload.AppID))
		err = steamHelper.AllowSteamCodes(err, b, []int{400})
		if err != nil {
			log.Err(err, message.Message.Body)
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

		timestamp, err := mongo.FindLatestPlayerAchievement(payload.PlayerID, payload.AppID)
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			continue
		}

		var rows []mongo.PlayerAchievement

		for _, v := range resp.Achievements {
			if v.Achieved {
				if v.UnlockTime >= timestamp {
					rows = append(rows, mongo.PlayerAchievement{
						PlayerID:               payload.PlayerID,
						AppID:                  payload.AppID,
						AchievementID:          v.APIName,
						AchievementName:        v.Name,
						AchievementDescription: v.Description,
						AchievementDate:        v.UnlockTime,
					})
				}
			}
		}

		err = mongo.UpdatePlayerAchievements(rows)
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
