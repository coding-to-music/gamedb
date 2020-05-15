package queue

import (
	"github.com/Jleagle/rabbit-go"
	elasticHelpers "github.com/gamedb/gamedb/pkg/elastic"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type GroupSearchMessage struct {
	Group mongo.Group `json:"group"`
}

func groupsSearchHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := GroupSearchMessage{}
		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		if payload.Group.Type != helpers.GroupTypeGroup {
			message.Ack(false)
			continue
		}

		err = elasticHelpers.IndexGroup(elasticHelpers.Group{
			ID:           payload.Group.ID,
			Name:         payload.Group.Name,
			URL:          payload.Group.URL,
			Abbreviation: payload.Group.Abbr,
			Headline:     payload.Group.Headline,
			Icon:         payload.Group.Icon,
			Members:      payload.Group.Members,
			Trend:        payload.Group.Trending,
		})
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
