package consumers

import (
	"github.com/gamedb/gamedb/pkg/consumers/framework"
	"github.com/gamedb/gamedb/pkg/log"
)

func bundleHandler(messages []framework.Message) {

	log.Info("bundle handler")

	for _, message := range messages {

		log.Info(message.Attempt())

		message.SendToQueue(queues[framework.Producer][queueApps])
	}
}
