package queue

import (
	"strconv"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/elastic"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
)

type AppsSearchMessage struct {
	ID        int                   `json:"id"`
	Name      string                `json:"name"`
	Icon      string                `json:"icon"`
	Players   int                   `json:"players"`
	Followers int                   `json:"folloers"`
	Score     float64               `json:"score"`
	Prices    helpers.ProductPrices `json:"prices"`
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

		client, ctx, err := elastic.GetElastic()
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToRetryQueue(message)
			continue
		}

		_, err = client.Index().
			Index(elastic.IndexApps).
			Id(strconv.Itoa(payload.ID)).
			BodyJson(payload).
			Do(ctx)

		if err != nil {
			log.Err(err, message.Message.Body)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
