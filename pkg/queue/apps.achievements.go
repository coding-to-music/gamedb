package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
)

type AppAchievementsMessage struct {
	ID int `json:"id"`
}

func appAchievementsHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := AppAchievementsMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		message.Ack(false)
	}
}
