package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/elastic"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type AppsAchievementsSearchMessage struct {
	AppAchievement mongo.AppAchievement `json:"app_achievement"`
	AppName        string               `json:"app_name"`
}

func appsAchievementsSearchHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := AppsAchievementsSearchMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		achievement := elastic.Achievement{}
		achievement.ID = payload.AppAchievement.Key
		achievement.AppID = payload.AppAchievement.AppID
		achievement.Name = payload.AppAchievement.Name
		achievement.Icon = payload.AppAchievement.Icon
		achievement.Description = payload.AppAchievement.Description
		achievement.Hidden = payload.AppAchievement.Hidden
		achievement.Completed = payload.AppAchievement.Completed
		achievement.AppName = payload.AppName

		if achievement.ID == "" || achievement.AppID == 0 {
			sendToFailQueue(message)
			continue
		}

		if payload.AppAchievement.Deleted {
			err = elastic.DeleteDocument(elastic.IndexAchievements, achievement.GetKey())
		} else {
			err = elastic.IndexAchievement(achievement)
		}
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
