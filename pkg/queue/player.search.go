package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/elastic"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type PlayersSearchMessage struct {
	Player mongo.Player `json:"player"`
}

func appsPlayersHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := PlayersSearchMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		player := elastic.Player{}
		player.ID = payload.Player.ID
		player.PersonaName = payload.Player.PersonaName
		player.PersonaNameRecent = []string{} // todo
		player.VanityURL = payload.Player.VanityURL

		err = elastic.IndexPlayer(player)
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
