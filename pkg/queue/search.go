package queue

import (
	"strconv"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/search"
	"github.com/gamedb/gamedb/pkg/log"
)

const (
	searchTypeApp    = "app"
	searchTypePlayer = "player"
)

type SearchMessage struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
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

		tweet1 := search.SearchResult{
			Keywords: []string{"james", "eagle"},
			Name:     "Jleagle",
			ID:       440,
			Icon:     "/test.png",
		}

		_, err = client.Index().
			Index("gdb-search").
			Id("app-" + strconv.FormatUint(440, 10)).
			BodyJson(tweet1).
			Do(ctx)

		if err != nil {
			log.Err(err, message.Message.Body)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
