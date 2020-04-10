package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
)

type AppSameownersMessage struct {
	ID int `json:"id"`
}

func appSameownersHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := AppSameownersMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		message.Ack(false)
	}
}
