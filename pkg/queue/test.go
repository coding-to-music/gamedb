package queue

import (
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type testMessage struct {
	baseMessage
	Message testMessageInner `json:"message"`
}

type testMessageInner struct {
	ID int `json:"id"`
}

type testQueue struct {
}

func (q testQueue) processMessages(msgs []amqp.Delivery) {

	msg := msgs[0]

	var err error

	message := testMessage{}
	message.OriginalQueue = QueueSteam

	err = helpers.Unmarshal(msg.Body, &message)
	if err != nil {
		log.Err(err, msg.Body)
		message.ack(msg)
		return
	}

	log.Info(message.Message.ID)

	message.ack(msg)
}
