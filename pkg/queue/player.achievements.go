package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
)

type PlayerAchievementsMessage struct {
	PlayerID int64 `json:"player_id"`
	AppID    int   `json:"app_id"`
}

func playerAchievementsHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := PlayerAchievementsMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		//resp, _, err := steamHelper.GetSteamUnlimited()

		message.Ack(false)
	}
}
