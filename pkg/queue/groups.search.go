package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.uber.org/zap"
)

type GroupSearchMessage struct {
	Group     *mongo.Group `json:"group"`
	GroupID   string       `json:"group_id"`   // Only if group is null
	GroupType string       `json:"group_type"` // Only if group is null
}

func (m GroupSearchMessage) Queue() rabbit.QueueName {
	return QueueGroupsSearch
}

func groupsSearchHandler(message *rabbit.Message) {

	payload := GroupSearchMessage{}
	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	if payload.GroupType != "" && payload.GroupType != helpers.GroupTypeGroup {
		message.Ack()
		return
	}

	var groupMongo mongo.Group

	if payload.GroupID != "" {

		groupMongo, err = mongo.GetGroup(payload.GroupID)
		if err != nil {
			log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
			sendToRetryQueue(message)
			return
		}

	} else if payload.Group != nil {

		groupMongo = *payload.Group

	} else {

		sendToFailQueue(message)
		return
	}

	if groupMongo.Type != helpers.GroupTypeGroup {
		message.Ack()
		return
	}

	group := elasticsearch.Group{
		ID:           groupMongo.ID,
		Name:         groupMongo.Name,
		URL:          groupMongo.URL,
		Abbreviation: groupMongo.Abbr,
		Headline:     groupMongo.Headline,
		Icon:         groupMongo.Icon,
		Members:      groupMongo.Members,
		Trend:        groupMongo.Trending,
		Error:        groupMongo.Error != "",
		Primaries:    groupMongo.Primaries,
	}

	err = elasticsearch.IndexGroup(group)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	message.Ack()
}
