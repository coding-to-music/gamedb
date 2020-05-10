package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/search"
	"github.com/gamedb/gamedb/pkg/log"
)

type SearchMessage struct {
	ID      uint64   `json:"id"`
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
	Type    string   `json:"type"`
	Icon    string   `json:"icon"`
}

func searchHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := SearchMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		client, ctx, err := search.GetElastic()
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToRetryQueue(message)
			continue
		}

		insert := search.SearchResult{}
		insert.ID = payload.ID
		insert.Name = payload.Name
		insert.Keywords = payload.Aliases
		insert.Type = payload.Type
		insert.Icon = insert.GetIcon()

		// log.Info("inserting: " + insert.GetKey())

		_, err = client.Index().
			Index(search.IndexName).
			Id(insert.GetKey()).
			BodyJson(insert).
			Do(ctx)

		if err != nil {
			log.Err(err, message.Message.Body)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
