package consumers

import (
	"github.com/gamedb/gamedb/pkg/consumers/framework"
)

func appHandler(messages []*framework.Message) {

	for _, message := range messages {
		message.SendToQueue(channels[framework.Producer][queueApps])
	}
}
