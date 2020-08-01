package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type GroupSearchMessage struct {
	Group mongo.Group `json:"group"`
}

func (m GroupSearchMessage) Queue() rabbit.QueueName {
	return QueueGroupsSearch
}

func groupsSearchHandler(message *rabbit.Message) {

	payload := GroupSearchMessage{}
	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err, message.Message.Body)
		sendToFailQueue(message)
		return
	}

	if payload.Group.Type != helpers.GroupTypeGroup {
		message.Ack(false)
		return
	}

	group := elasticsearch.Group{
		ID:           payload.Group.ID,
		Name:         payload.Group.Name,
		URL:          payload.Group.URL,
		Abbreviation: payload.Group.Abbr,
		Headline:     payload.Group.Headline,
		Icon:         payload.Group.Icon,
		Members:      payload.Group.Members,
		Trend:        payload.Group.Trending,
		Error:        payload.Group.Error != "",
		Primaries:    payload.Group.Primaries,
	}

	err = elasticsearch.IndexGroup(group)
	if err != nil {
		log.Err(err, message.Message.Body)
		sendToRetryQueue(message)
		return
	}

	message.Ack(false)
}
