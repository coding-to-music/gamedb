package queue

import (
	"github.com/Jleagle/rabbit-go"
)

type AppsPlayersMessage struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

func appsPlayersHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		message.Ack(false)
	}
}
