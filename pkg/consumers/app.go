package consumers

import (
	"github.com/gamedb/gamedb/pkg/consumers/framework"
	"github.com/gamedb/gamedb/pkg/log"
)

func appHandler(messages []framework.Message) {

	log.Info("app handler")

	for _, message := range messages {

		log.Info(message.Attempt())

		if message.Attempt() > 5 {
			message.Ack()
		} else {
			message.SendToQueue(queues[framework.Producer][queueBundles])
		}
	}
}
