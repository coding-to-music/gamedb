package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/search"
	"github.com/gamedb/gamedb/pkg/log"
)

type AppsSearchMessage struct {
	ID      uint64   `json:"id"`
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
	Icon    string   `json:"icon"`
}

func appsSearchHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := AppsSearchMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		if payload.Name == "" || payload.ID == 0 || payload.Type == "" {
			message.Ack(false)
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
