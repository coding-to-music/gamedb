package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type PlayersSearchMessage struct {
	Player mongo.Player `json:"player"`
}

func appsPlayersHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		message.Ack(false)
	}
}
