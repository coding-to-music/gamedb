package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/elastic"
	"github.com/gamedb/gamedb/pkg/log"
)

const (
	searchTypeApp    = "app"
	searchTypePlayer = "player"
)

type searchMessage struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func searchHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := searchMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		client, err := elastic.GetElastic()

		log.Info(client.Info())

		message.Ack(false)
	}
}
