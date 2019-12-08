package queue

import (
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type groupMessage struct {
	baseMessage
	Message groupMessageInner `json:"message"`
}

type groupMessageInner struct {
	IDs []string `json:"ids"`
}

type groupQueueScrape struct {
}

func (q groupQueueScrape) processMessages(msgs []amqp.Delivery) {

	message := groupMessage{}

	err := helpers.Unmarshal(msgs[0].Body, &message)
	if err != nil {
		log.Err(err, msgs[0].Body)
		ackFail(msgs[0], &message)
		return
	}

	payload := consumers.GroupMessage{}
	payload.IDs = message.Message.IDs

	err = consumers.ProduceGroup(payload)
	if err != nil {
		log.Err(err, msgs[0].Body)
		ackRetry(msgs[0], &message)
	} else {
		message.ack(msgs[0])
	}
}
