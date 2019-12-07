package consumers

import (
	"github.com/gamedb/gamedb/pkg/consumers/framework"
)

type AppMessage struct {
	ID           int                    `json:"id"`
	ChangeNumber int                    `json:"change_number"`
	VDF          map[string]interface{} `json:"vdf"`
}

func appHandler(messages []*framework.Message) {

	for _, message := range messages {
		message.SendToQueue(channels[framework.Producer][queueApps])
	}
}
